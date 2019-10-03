package svc

import (
	"encoding/json"
	"github.com/dgraph-io/badger"
	"lyra/restApiV1"
	"lyra/tool"
	"reflect"
	"sort"
	"strings"
	"time"
)

func (s *Service) ReadPlaylists(externalTrn *badger.Txn, filter *restApiV1.PlaylistFilter) ([]*restApiV1.Playlist, error) {
	playlists := []*restApiV1.Playlist{}

	opts := badger.DefaultIteratorOptions
	switch filter.Order {
	case restApiV1.PlaylistOrderByPlaylistName:
		opts.Prefix = []byte(playlistNamePlaylistIdPrefix)
		opts.PrefetchValues = false
	case restApiV1.PlaylistOrderByUpdateTs:
		opts.Prefix = []byte(playlistUpdateTsPlaylistIdPrefix)
		opts.PrefetchValues = false
	case restApiV1.PlaylistOrderByContentUpdateTs:
		opts.Prefix = []byte(playlistContentUpdateTsPlaylistIdPrefix)
		opts.PrefetchValues = false
	default:
		opts.Prefix = []byte(playlistIdPrefix)
	}

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	if filter.Order == restApiV1.PlaylistOrderByUpdateTs {
		it.Seek([]byte(playlistUpdateTsPlaylistIdPrefix + indexTs(filter.FromTs)))
	} else if filter.Order == restApiV1.PlaylistOrderByContentUpdateTs {
		it.Seek([]byte(playlistContentUpdateTsPlaylistIdPrefix + indexTs(filter.FromTs)))
	} else {
		it.Rewind()
	}

	for ; it.Valid(); it.Next() {
		var playlist *restApiV1.Playlist

		switch filter.Order {
		case restApiV1.PlaylistOrderByPlaylistName,
			restApiV1.PlaylistOrderByUpdateTs,
			restApiV1.PlaylistOrderByContentUpdateTs:
			key := it.Item().KeyCopy(nil)

			playlistId := strings.Split(string(key), ":")[2]
			var e error
			playlist, e = s.ReadPlaylist(txn, playlistId)
			if e != nil {
				return nil, e
			}
		default:
			encodedPlaylist, e := it.Item().ValueCopy(nil)
			if e != nil {
				return nil, e
			}
			e = json.Unmarshal(encodedPlaylist, &playlist)
			if e != nil {
				return nil, e
			}
		}

		playlists = append(playlists, playlist)

	}

	return playlists, nil
}

func (s *Service) ReadPlaylist(externalTrn *badger.Txn, playlistId string) (*restApiV1.Playlist, error) {
	var playlist *restApiV1.Playlist

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	item, e := txn.Get(getPlaylistIdKey(playlistId))
	if e != nil {
		if e == badger.ErrKeyNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}
	encodedPlaylist, e := item.ValueCopy(nil)
	if e != nil {
		return nil, e
	}
	e = json.Unmarshal(encodedPlaylist, &playlist)
	if e != nil {
		return nil, e
	}

	return playlist, nil
}

func (s *Service) CreatePlaylist(externalTrn *badger.Txn, playlistMeta *restApiV1.PlaylistMeta) (*restApiV1.Playlist, error) {
	return s.CreateInternalPlaylist(externalTrn, "", playlistMeta)
}

func (s *Service) CreateInternalPlaylist(externalTrn *badger.Txn, playlistId string, playlistMeta *restApiV1.PlaylistMeta) (*restApiV1.Playlist, error) {
	var playlist *restApiV1.Playlist

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	// Create playlist
	now := time.Now().UnixNano()

	if playlistId == "" {
		playlistId = tool.CreateUlid()
	}

	playlist = &restApiV1.Playlist{
		Id:              playlistId,
		CreationTs:      now,
		UpdateTs:        now,
		ContentUpdateTs: now,
		PlaylistMeta:    *playlistMeta,
	}

	// Clean owner list
	playlist.OwnerUserIds = tool.Deduplicate(playlist.OwnerUserIds)
	sort.Slice(playlist.OwnerUserIds, func(i, j int) bool {
		return playlist.OwnerUserIds[i] < playlist.OwnerUserIds[j]
	})

	encodedPlaylist, _ := json.Marshal(playlist)
	e := txn.Set(getPlaylistIdKey(playlist.Id), encodedPlaylist)
	if e != nil {
		return nil, e
	}

	// Create playlistName Index
	e = txn.Set(getPlaylistNamePlaylistIdKey(playlist.Name, playlist.Id), nil)
	if e != nil {
		return nil, e
	}

	// Create playlist updateTs index
	e = txn.Set(getPlaylistUpdateTsPlaylistIdKey(playlist.UpdateTs, playlist.Id), nil)
	if e != nil {
		return nil, e
	}

	// Create playlist contentUpdateTs index
	e = txn.Set(getPlaylistContentUpdateTsPlaylistIdKey(playlist.ContentUpdateTs, playlist.Id), nil)
	if e != nil {
		return nil, e
	}

	// Create owners link
	for _, ownerUserId := range playlist.OwnerUserIds {
		// Check owner user id
		_, e = txn.Get(getUserIdKey(ownerUserId))
		if e != nil {
			return nil, e
		}

		// Store playlist owner
		e = txn.Set(getUserIdOwnedPlaylistIdKey(ownerUserId, playlistId), nil)
		if e != nil {
			return nil, e
		}
	}

	// Create songs link
	for _, songId := range playlist.SongIds {
		// Check song id
		_, e := txn.Get(getSongIdKey(songId))
		if e != nil {
			return nil, e
		}

		// Store song link
		e = txn.Set(getSongIdPlaylistIdKey(songId, playlist.Id), nil)
		if e != nil {
			return nil, e
		}
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return playlist, nil
}

func (s *Service) UpdatePlaylist(externalTrn *badger.Txn, playlistId string, playlistMeta *restApiV1.PlaylistMeta) (*restApiV1.Playlist, error) {

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	now := time.Now().UnixNano()

	playlist, e := s.ReadPlaylist(txn, playlistId)

	if e != nil {
		return nil, e
	}

	playlistOldName := playlist.Name
	playlistOldUpdateTs := playlist.UpdateTs
	playlistOldContentUpdateTs := playlist.ContentUpdateTs
	playlistOldSongIds := playlist.PlaylistMeta.SongIds
	playlistOldOwnerUserIds := playlist.PlaylistMeta.OwnerUserIds

	if playlistMeta != nil {
		playlist.PlaylistMeta = *playlistMeta
	}

	// Clean owner list
	playlist.OwnerUserIds = tool.Deduplicate(playlist.OwnerUserIds)
	sort.Slice(playlist.OwnerUserIds, func(i, j int) bool {
		return playlist.OwnerUserIds[i] < playlist.OwnerUserIds[j]
	})

	// Detect song list update
	songIdsUpdated := !reflect.DeepEqual(playlistOldSongIds, playlist.PlaylistMeta.SongIds)

	// Update playlist update timestamp
	playlist.UpdateTs = now

	e = tool.ReplaceKey(
		txn,
		getPlaylistUpdateTsPlaylistIdKey(playlistOldUpdateTs, playlist.Id),
		getPlaylistUpdateTsPlaylistIdKey(playlist.UpdateTs, playlist.Id),
	)

	// Update playlist update content timestamp
	if playlistOldName != playlist.Name || songIdsUpdated {
		playlist.ContentUpdateTs = now

		// Update playlist contentUpdateTs index
		e = tool.ReplaceKey(
			txn,
			getPlaylistContentUpdateTsPlaylistIdKey(playlistOldContentUpdateTs, playlist.Id),
			getPlaylistContentUpdateTsPlaylistIdKey(playlist.ContentUpdateTs, playlist.Id),
		)

	}

	encodedPlaylist, _ := json.Marshal(playlist)
	e = txn.Set(getPlaylistIdKey(playlist.Id), encodedPlaylist)
	if e != nil {
		return nil, e
	}

	// Update playlist name index
	if playlistOldName != playlist.Name {
		e = tool.ReplaceKey(
			txn,
			getPlaylistNamePlaylistIdKey(playlistOldName, playlist.Id),
			getPlaylistNamePlaylistIdKey(playlist.Name, playlist.Id),
		)
	}

	// Update owner index
	for _, ownerUserId := range playlistOldOwnerUserIds {
		e = txn.Delete(getUserIdOwnedPlaylistIdKey(ownerUserId, playlistId))
		if e != nil {
			return nil, e
		}
	}
	for _, ownerUserId := range playlist.OwnerUserIds {
		// Check owner user id
		_, e := txn.Get(getUserIdKey(ownerUserId))
		if e != nil {
			return nil, e
		}

		// Store playlist owner
		e = txn.Set(getUserIdOwnedPlaylistIdKey(ownerUserId, playlistId), nil)
		if e != nil {
			return nil, e
		}
	}

	// Update songs list
	if songIdsUpdated {
		for _, songId := range playlistOldSongIds {
			e = txn.Delete(getSongIdPlaylistIdKey(songId, playlist.Id))
			if e != nil {
				return nil, e
			}
		}
		for _, songId := range playlist.SongIds {
			// Check song id
			_, e := txn.Get(getSongIdKey(songId))
			if e != nil {
				return nil, e
			}

			// Store song link
			e = txn.Set(getSongIdPlaylistIdKey(songId, playlist.Id), nil)
			if e != nil {
				return nil, e
			}
		}
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return playlist, nil
}

func (s *Service) DeletePlaylist(externalTrn *badger.Txn, playlistId string) (*restApiV1.Playlist, error) {
	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	deleteTs := time.Now().UnixNano()

	playlist, e := s.ReadPlaylist(txn, playlistId)
	if e != nil {
		return nil, e
	}

	// Delete playlist name index
	e = txn.Delete(getPlaylistNamePlaylistIdKey(playlist.Name, playlistId))
	if e != nil {
		return nil, e
	}

	// Delete ower link
	for _, owenrUserId := range playlist.OwnerUserIds {
		e = txn.Delete(getUserIdOwnedPlaylistIdKey(owenrUserId, playlistId))
		if e != nil {
			return nil, e
		}
	}

	// Delete songs link
	for _, songId := range playlist.SongIds {
		e = txn.Delete(getSongIdPlaylistIdKey(songId, playlist.Id))
		if e != nil {
			return nil, e
		}
	}

	// Delete updateTs index
	e = txn.Delete(getPlaylistUpdateTsPlaylistIdKey(playlist.UpdateTs, playlist.Id))
	if e != nil {
		return nil, e
	}

	// Delete contentUpdateTs index
	e = txn.Delete(getPlaylistContentUpdateTsPlaylistIdKey(playlist.ContentUpdateTs, playlist.Id))
	if e != nil {
		return nil, e
	}

	// Delete playlist
	e = txn.Delete(getPlaylistIdKey(playlistId))
	if e != nil {
		return nil, e
	}

	// Archive playlistId
	e = txn.Set(getPlaylistDeleteTsPlaylistIdKey(deleteTs, playlist.Id), nil)
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return playlist, nil
}

func (s *Service) GetDeletedPlaylistIds(externalTrn *badger.Txn, fromTs int64) ([]string, error) {

	playlistIds := []string{}

	opts := badger.DefaultIteratorOptions
	opts.Prefix = []byte(playlistDeleteTsPlaylistIdPrefix)
	opts.PrefetchValues = false

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	for it.Seek([]byte(playlistDeleteTsPlaylistIdPrefix + indexTs(fromTs))); it.Valid(); it.Next() {

		key := it.Item().KeyCopy(nil)

		playlistId := strings.Split(string(key), ":")[2]

		playlistIds = append(playlistIds, playlistId)

	}

	return playlistIds, nil
}

func (s *Service) GetPlaylistIdsByName(externalTrn *badger.Txn, playlistName string) ([]string, error) {
	var playlistIds []string

	opts := badger.DefaultIteratorOptions
	opts.Prefix = []byte(playlistNamePlaylistIdPrefix + indexString(playlistName) + ":")
	opts.PrefetchValues = false

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	for it.Rewind(); it.Valid(); it.Next() {

		key := it.Item().KeyCopy(nil)

		playlistId := strings.Split(string(key), ":")[2]

		playlistIds = append(playlistIds, playlistId)

	}

	return playlistIds, nil
}

func (s *Service) GetPlaylistIdsFromSongId(externalTrn *badger.Txn, songId string) ([]string, error) {

	var playlistIds []string

	opts := badger.DefaultIteratorOptions
	opts.Prefix = []byte(songIdPlaylistIdPrefix + songId + ":")
	opts.PrefetchValues = false

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	for it.Rewind(); it.Valid(); it.Next() {

		key := it.Item().KeyCopy(nil)

		playlistId := strings.Split(string(key), ":")[2]

		playlistIds = append(playlistIds, playlistId)

	}

	return playlistIds, nil

}

func (s *Service) GetPlaylistIdsFromOwnerUserId(externalTrn *badger.Txn, ownerUserId string) ([]string, error) {

	var playlistIds []string

	opts := badger.DefaultIteratorOptions
	opts.Prefix = []byte(userIdOwnedPlaylistIdPrefix + ownerUserId + ":")
	opts.PrefetchValues = false

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	for it.Rewind(); it.Valid(); it.Next() {

		key := it.Item().KeyCopy(nil)

		playlistId := strings.Split(string(key), ":")[2]

		playlistIds = append(playlistIds, playlistId)

	}

	return playlistIds, nil

}
