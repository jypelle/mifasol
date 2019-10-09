package svc

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"lyra/restApiV1"
	"lyra/tool"
	"reflect"
	"sort"
	"time"
)

func (s *Service) ReadPlaylists(externalTrn storm.Node, filter *restApiV1.PlaylistFilter) ([]restApiV1.Playlist, error) {
	playlists := []restApiV1.Playlist{}

	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(false)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	var matchers []q.Matcher

	if filter.FromTs != nil {
		matchers = append(matchers, q.Gte("UpdateTs", *filter.FromTs))
	}
	if filter.ContentFromTs != nil {
		matchers = append(matchers, q.Gte("ContentUpdateTs", *filter.ContentFromTs))
	}

	query := txn.Select(matchers...)

	switch filter.Order {
	case restApiV1.PlaylistOrderByPlaylistName:
		query = query.OrderBy("Name")
	case restApiV1.PlaylistOrderByUpdateTs:
		query = query.OrderBy("UpdateTs")
	case restApiV1.PlaylistOrderByContentUpdateTs:
		query = query.OrderBy("ContentUpdateTs")
	default:
		query = query.OrderBy("Id")
	}

	err = query.Find(&playlists)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	return playlists, nil
}

func (s *Service) ReadPlaylist(externalTrn storm.Node, playlistId string) (*restApiV1.Playlist, error) {
	var playlist restApiV1.Playlist

	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(false)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	e := txn.One("Id", playlistId, &playlist)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}

	return &playlist, nil
}

func (s *Service) CreatePlaylist(externalTrn storm.Node, playlistMeta *restApiV1.PlaylistMeta) (*restApiV1.Playlist, error) {
	return s.CreateInternalPlaylist(externalTrn, "", playlistMeta)
}

func (s *Service) CreateInternalPlaylist(externalTrn storm.Node, playlistId string, playlistMeta *restApiV1.PlaylistMeta) (*restApiV1.Playlist, error) {

	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(true)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	// Create playlist
	now := time.Now().UnixNano()

	if playlistId == "" {
		playlistId = tool.CreateUlid()
	}

	playlist := &restApiV1.Playlist{
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

	e := txn.Save(playlist)
	if e != nil {
		return nil, e
	}

	// Create owners link
	for _, ownerUserId := range playlist.OwnerUserIds {
		// Check owner user id
		var userComplete restApiV1.UserComplete
		e := txn.One("Id", ownerUserId, &userComplete)
		if e != nil {
			return nil, e
		}

		// Store playlist owner
		e = txn.Save(&restApiV1.OwnedUserPlaylist{UserId: ownerUserId, PlaylistId: playlistId})
		if e != nil {
			return nil, e
		}
	}

	// Create songs link
	for _, songId := range playlist.SongIds {
		// Check song id
		var song restApiV1.Song
		e := txn.One("Id", songId, &song)
		if e != nil {
			return nil, e
		}

		// Store song link
		e = txn.Save(&restApiV1.PlaylistSong{PlaylistId: playlistId, SongId: songId})
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

func (s *Service) UpdatePlaylist(externalTrn storm.Node, playlistId string, playlistMeta *restApiV1.PlaylistMeta) (*restApiV1.Playlist, error) {

	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(true)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	now := time.Now().UnixNano()

	playlist, e := s.ReadPlaylist(txn, playlistId)

	if e != nil {
		return nil, e
	}

	playlistOldName := playlist.Name
	playlistOldSongIds := playlist.PlaylistMeta.SongIds

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

	// Update playlist update content timestamp
	if playlistOldName != playlist.Name || songIdsUpdated {
		playlist.ContentUpdateTs = now
	}

	e = txn.Update(playlist)
	if e != nil {
		return nil, e
	}

	// Update owner index
	query := txn.Select(q.Eq("PlaylistId", playlistId))
	e = query.Delete(new(restApiV1.OwnedUserPlaylist))
	if e != nil {
		return nil, e
	}

	for _, ownerUserId := range playlist.OwnerUserIds {
		// Check owner user id
		var userComplete restApiV1.UserComplete
		e := txn.One("Id", ownerUserId, &userComplete)
		if e != nil {
			return nil, e
		}

		// Store playlist owner
		e = txn.Save(&restApiV1.OwnedUserPlaylist{UserId: ownerUserId, PlaylistId: playlistId})
		if e != nil {
			return nil, e
		}
	}

	// Update songs list
	if songIdsUpdated {
		query := txn.Select(q.Eq("PlaylistId", playlistId))
		e = query.Delete(new(restApiV1.PlaylistSong))
		if e != nil {
			return nil, e
		}
		for _, songId := range playlist.SongIds {
			// Check song id
			var song restApiV1.Song
			e := txn.One("Id", songId, &song)
			if e != nil {
				return nil, e
			}

			// Store song link
			e = txn.Save(&restApiV1.PlaylistSong{PlaylistId: playlistId, SongId: songId})
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

func (s *Service) DeletePlaylist(externalTrn storm.Node, playlistId string) (*restApiV1.Playlist, error) {
	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(true)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	deleteTs := time.Now().UnixNano()

	playlist, e := s.ReadPlaylist(txn, playlistId)
	if e != nil {
		return nil, e
	}

	// Delete ower link
	query := txn.Select(q.Eq("PlaylistId", playlistId))
	e = query.Delete(new(restApiV1.OwnedUserPlaylist))
	if e != nil {
		return nil, e
	}

	// Delete songs link
	query = txn.Select(q.Eq("PlaylistId", playlistId))
	e = query.Delete(new(restApiV1.PlaylistSong))
	if e != nil {
		return nil, e
	}

	// Delete playlist
	e = txn.DeleteStruct(playlist)
	if e != nil {
		return nil, e
	}

	// Archive playlistId
	e = txn.Save(&restApiV1.DeletedPlaylist{Id: playlist.Id, DeleteTs: deleteTs})
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return playlist, nil
}

func (s *Service) GetDeletedPlaylistIds(externalTrn storm.Node, fromTs int64) ([]string, error) {

	playlistIds := []string{}
	deletedPlaylists := []restApiV1.DeletedPlaylist{}

	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(false)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	query := txn.Select(q.Gte("DeleteTs", fromTs)).OrderBy("DeleteTs")

	err = query.Find(&deletedPlaylists)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	for _, deletedPlaylist := range deletedPlaylists {
		playlistIds = append(playlistIds, deletedPlaylist.Id)
	}

	return playlistIds, nil
}

func (s *Service) GetPlaylistIdsFromSongId(externalTrn storm.Node, songId string) ([]string, error) {

	var playlistIds []string
	playlistSongs := []restApiV1.PlaylistSong{}

	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(false)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	query := txn.Select(q.Eq("SongId", songId))

	err = query.Find(&playlistSongs)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	for _, playlistSong := range playlistSongs {
		playlistIds = append(playlistIds, playlistSong.PlaylistId)
	}

	return playlistIds, nil

}

func (s *Service) GetPlaylistIdsFromOwnerUserId(externalTrn storm.Node, ownerUserId string) ([]string, error) {

	var playlistIds []string
	ownedUserPlaylists := []restApiV1.OwnedUserPlaylist{}

	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(false)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	query := txn.Select(q.Eq("UserId", ownerUserId))
	err = query.Find(&ownedUserPlaylists)

	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	for _, ownedUserPlaylist := range ownedUserPlaylists {
		playlistIds = append(playlistIds, ownedUserPlaylist.PlaylistId)
	}

	return playlistIds, nil

}
