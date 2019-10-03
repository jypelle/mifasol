package svc

import (
	"encoding/json"
	"github.com/dgraph-io/badger"
	"lyra/restApiV1"
	"lyra/tool"
	"sort"
	"strings"
	"time"
)

func (s *Service) ReadAlbums(externalTrn *badger.Txn, filter *restApiV1.AlbumFilter) ([]*restApiV1.Album, error) {
	albums := []*restApiV1.Album{}

	opts := badger.DefaultIteratorOptions
	switch filter.Order {
	case restApiV1.AlbumOrderByAlbumName:
		opts.Prefix = []byte(albumNameAlbumIdPrefix)
		opts.PrefetchValues = false
	case restApiV1.AlbumOrderByUpdateTs:
		opts.Prefix = []byte(albumUpdateTsAlbumIdPrefix)
		opts.PrefetchValues = false
	default:
		opts.Prefix = []byte(albumIdPrefix)
	}

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	if filter.Order == restApiV1.AlbumOrderByUpdateTs {
		it.Seek([]byte(albumUpdateTsAlbumIdPrefix + indexTs(filter.FromTs)))
	} else {
		it.Rewind()
	}

	for ; it.Valid(); it.Next() {
		var album *restApiV1.Album

		switch filter.Order {
		case restApiV1.AlbumOrderByAlbumName,
			restApiV1.AlbumOrderByUpdateTs:
			key := it.Item().KeyCopy(nil)

			albumId := strings.Split(string(key), ":")[2]
			var e error
			album, e = s.ReadAlbum(txn, albumId)
			if e != nil {
				return nil, e
			}
		default:
			encodedAlbum, e := it.Item().ValueCopy(nil)
			if e != nil {
				return nil, e
			}
			e = json.Unmarshal(encodedAlbum, &album)
			if e != nil {
				return nil, e
			}
		}

		albums = append(albums, album)

	}

	return albums, nil
}

func (s *Service) ReadAlbum(externalTrn *badger.Txn, albumId string) (*restApiV1.Album, error) {
	var album *restApiV1.Album

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	item, e := txn.Get(getAlbumIdKey(albumId))
	if e != nil {
		if e == badger.ErrKeyNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}
	encodedAlbum, e := item.ValueCopy(nil)
	if e != nil {
		return nil, e
	}
	e = json.Unmarshal(encodedAlbum, &album)
	if e != nil {
		return nil, e
	}

	return album, nil
}

func (s *Service) CreateAlbum(externalTrn *badger.Txn, albumMeta *restApiV1.AlbumMeta) (*restApiV1.Album, error) {
	var album *restApiV1.Album

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	// Store album
	now := time.Now().UnixNano()

	album = &restApiV1.Album{
		Id:         tool.CreateUlid(),
		CreationTs: now,
		UpdateTs:   now,
		AlbumMeta:  *albumMeta,
	}

	encodedAlbum, _ := json.Marshal(album)
	e := txn.Set(getAlbumIdKey(album.Id), encodedAlbum)
	if e != nil {
		return nil, e
	}
	// Store albumName Index
	e = txn.Set(getAlbumNameAlbumIdKey(album.Name, album.Id), nil)
	if e != nil {
		return nil, e
	}

	// Store updateTs Index
	e = txn.Set(getAlbumUpdateTsAlbumIdKey(album.UpdateTs, album.Id), nil)
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return album, nil
}

func (s *Service) UpdateAlbum(externalTrn *badger.Txn, albumId string, albumMeta *restApiV1.AlbumMeta) (*restApiV1.Album, error) {
	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	album, err := s.ReadAlbum(txn, albumId)
	if err != nil {
		return nil, err
	}

	albumOldName := album.Name
	albumOldUpdateTs := album.UpdateTs
	album.AlbumMeta = *albumMeta

	// Update album
	album.UpdateTs = time.Now().UnixNano()
	encodedAlbum, _ := json.Marshal(album)
	e := txn.Set(getAlbumIdKey(album.Id), encodedAlbum)
	if e != nil {
		return nil, e
	}

	// Update album name Index
	e = txn.Delete(getAlbumNameAlbumIdKey(albumOldName, album.Id))
	if e != nil {
		return nil, e
	}
	e = txn.Set(getAlbumNameAlbumIdKey(album.Name, album.Id), nil)
	if e != nil {
		return nil, e
	}

	// Update updateTs Index
	e = txn.Delete(getAlbumUpdateTsAlbumIdKey(albumOldUpdateTs, album.Id))
	if e != nil {
		return nil, e
	}

	e = txn.Set(getAlbumUpdateTsAlbumIdKey(album.UpdateTs, album.Id), nil)
	if e != nil {
		return nil, e
	}

	// Update tags in songs content
	songIds, e := s.GetSongIdsFromAlbumId(txn, albumId)
	if e != nil {
		return nil, e
	}

	for _, songId := range songIds {
		s.UpdateSong(txn, songId, nil, nil)
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return album, nil
}

func (s *Service) refreshAlbumArtistIds(externalTrn *badger.Txn, albumId string, updateArtistMetaArtistId *string) error {
	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	album, e := s.ReadAlbum(txn, albumId)
	if e != nil {
		return e
	}

	songIds, e := s.GetSongIdsFromAlbumId(txn, albumId)
	if e != nil {
		return e
	}

	albumOldUpdateTs := album.UpdateTs
	albumOldArtistIds := album.ArtistIds

	// Update AlbumArtists
	artistsCount := make(map[string]int)
	for _, songId := range songIds {

		song, e := s.ReadSong(txn, songId)
		if e != nil {
			return e
		}

		for _, artistId := range song.ArtistIds {
			if val, ok := artistsCount[artistId]; ok {
				artistsCount[artistId] = val + 1
			} else {
				artistsCount[artistId] = 1
			}
		}
	}

	album.ArtistIds = []string{}

	for artistId, artistCount := range artistsCount {
		if artistCount > len(songIds)/2 {
			album.ArtistIds = append(album.ArtistIds, artistId)
		}
	}

	// Reorder artists
	sort.Slice(album.ArtistIds, func(i, j int) bool {
		artistI, _ := s.ReadArtist(txn, album.ArtistIds[i])
		artistJ, _ := s.ReadArtist(txn, album.ArtistIds[j])
		return artistI.Name < artistJ.Name
	})

	artistIdsChanged := !isArtistIdsEqual(albumOldArtistIds, album.ArtistIds)
	isUpdatedArtistMetaInAlbumArtistIds := false
	if updateArtistMetaArtistId != nil {
		for _, artistId := range album.ArtistIds {
			if artistId == *updateArtistMetaArtistId {
				isUpdatedArtistMetaInAlbumArtistIds = true
				break
			}
		}
	}

	if artistIdsChanged || isUpdatedArtistMetaInAlbumArtistIds {
		// Update Song AlbumArtists
		for _, songId := range songIds {
			s.updateSongAlbumArtists(txn, songId, album.ArtistIds)
		}

		// Update album
		album.UpdateTs = time.Now().UnixNano()
		encodedAlbum, _ := json.Marshal(album)
		e = txn.Set(getAlbumIdKey(album.Id), encodedAlbum)
		if e != nil {
			return e
		}

		// Update updateTs Index
		e = txn.Delete(getAlbumUpdateTsAlbumIdKey(albumOldUpdateTs, album.Id))
		if e != nil {
			return e
		}

		e = txn.Set(getAlbumUpdateTsAlbumIdKey(album.UpdateTs, album.Id), nil)
		if e != nil {
			return e
		}

		// Commit transaction
		if externalTrn == nil {
			txn.Commit()
		}
	}

	return nil
}

func (s *Service) DeleteAlbum(externalTrn *badger.Txn, albumId string) (*restApiV1.Album, error) {
	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	deleteTs := time.Now().UnixNano()

	album, e := s.ReadAlbum(txn, albumId)

	if e != nil {
		return nil, e
	}

	// Check songs link
	songIds, e := s.GetSongIdsFromAlbumId(txn, albumId)
	if e != nil {
		return nil, e
	}
	if len(songIds) > 0 {
		return nil, ErrDeleteAlbumWithSongs
	}

	// Delete album name index
	e = txn.Delete(getAlbumNameAlbumIdKey(album.Name, albumId))
	if e != nil {
		return nil, e
	}

	// Delete album updateTs index
	e = txn.Delete(getAlbumUpdateTsAlbumIdKey(album.UpdateTs, albumId))
	if e != nil {
		return nil, e
	}

	// Delete album
	e = txn.Delete(getAlbumIdKey(albumId))
	if e != nil {
		return nil, e
	}

	// Archive albumId
	e = txn.Set(getAlbumDeleteTsAlbumIdKey(deleteTs, album.Id), nil)
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return album, nil
}

func (s *Service) GetDeletedAlbumIds(externalTrn *badger.Txn, fromTs int64) ([]string, error) {

	albumIds := []string{}

	opts := badger.DefaultIteratorOptions
	opts.Prefix = []byte(albumDeleteTsAlbumIdPrefix)
	opts.PrefetchValues = false

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	for it.Seek([]byte(albumDeleteTsAlbumIdPrefix + indexTs(fromTs))); it.Valid(); it.Next() {

		key := it.Item().KeyCopy(nil)

		albumId := strings.Split(string(key), ":")[2]

		albumIds = append(albumIds, albumId)

	}

	return albumIds, nil
}

func (s *Service) GetAlbumIdsByName(externalTrn *badger.Txn, albumName string) ([]string, error) {
	var albumIds []string

	opts := badger.DefaultIteratorOptions
	opts.Prefix = []byte(albumNameAlbumIdPrefix + indexString(albumName) + ":")
	opts.PrefetchValues = false

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	for it.Rewind(); it.Valid(); it.Next() {

		key := it.Item().KeyCopy(nil)

		albumId := strings.Split(string(key), ":")[2]

		albumIds = append(albumIds, albumId)

	}

	return albumIds, nil
}
