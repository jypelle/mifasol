package svc

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/srv/entity"
	"github.com/jypelle/mifasol/tool"
	"time"
)

func (s *Service) ReadAlbums(externalTrn storm.Node, filter *restApiV1.AlbumFilter) ([]restApiV1.Album, error) {
	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(false)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	var matchers []q.Matcher

	if filter.FromTs != nil {
		matchers = append(matchers, q.Gte("UpdateTs", *filter.FromTs))
	}
	if filter.Name != nil {
		matchers = append(matchers, q.Eq("Name", *filter.Name))
	}

	query := txn.Select(matchers...)

	switch filter.Order {
	case restApiV1.AlbumOrderByAlbumName:
		query = query.OrderBy("Name")
	case restApiV1.AlbumOrderByUpdateTs:
		query = query.OrderBy("UpdateTs")
	default:
	}

	albumEntities := []entity.AlbumEntity{}
	e = query.Find(&albumEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	albums := []restApiV1.Album{}

	for _, albumEntity := range albumEntities {
		var album restApiV1.Album
		albumEntity.Fill(&album)
		albums = append(albums, album)
	}

	return albums, nil
}

func (s *Service) ReadAlbum(externalTrn storm.Node, albumId restApiV1.AlbumId) (*restApiV1.Album, error) {
	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(false)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	var albumEntity entity.AlbumEntity
	e = txn.One("Id", albumId, &albumEntity)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}

	var album restApiV1.Album
	albumEntity.Fill(&album)

	return &album, nil
}

func (s *Service) CreateAlbum(externalTrn storm.Node, albumMeta *restApiV1.AlbumMeta) (*restApiV1.Album, error) {
	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(true)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	// Store album
	now := time.Now().UnixNano()

	albumEntity := entity.AlbumEntity{
		Id:         restApiV1.AlbumId(tool.CreateUlid()),
		CreationTs: now,
		UpdateTs:   now,
	}
	albumEntity.LoadMeta(albumMeta)

	e = txn.Save(&albumEntity)
	if e != nil {
		return nil, e
	}

	var album restApiV1.Album
	albumEntity.Fill(&album)

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return &album, nil
}

func (s *Service) UpdateAlbum(externalTrn storm.Node, albumId restApiV1.AlbumId, albumMeta *restApiV1.AlbumMeta) (*restApiV1.Album, error) {
	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(true)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	var albumEntity entity.AlbumEntity
	e = txn.One("Id", albumId, &albumEntity)
	if e != nil {
		return nil, e
	}

	oldName := albumEntity.Name

	albumEntity.LoadMeta(albumMeta)
	albumEntity.UpdateTs = time.Now().UnixNano()

	// Update album
	e = txn.Update(&albumEntity)
	if e != nil {
		return nil, e
	}

	// Update tags in songs content
	if oldName != albumEntity.Name {
		songIds, e := s.GetSongIdsFromAlbumId(txn, albumId)
		if e != nil {
			return nil, e
		}

		for _, songId := range songIds {
			s.UpdateSong(txn, songId, nil, nil, false)
		}
	}

	var album restApiV1.Album
	albumEntity.Fill(&album)

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return &album, nil
}

func (s *Service) refreshAlbumArtistIds(externalTrn storm.Node, albumId restApiV1.AlbumId, updateArtistMetaArtistId *restApiV1.ArtistId) error {
	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(true)
		if e != nil {
			return e
		}
		defer txn.Rollback()
	}

	var albumEntity entity.AlbumEntity
	e = txn.One("Id", albumId, &albumEntity)
	if e != nil {
		return e
	}

	songIds, e := s.GetSongIdsFromAlbumId(txn, albumId)
	if e != nil {
		return e
	}

	albumOldArtistIds := albumEntity.ArtistIds

	// Update AlbumArtists
	artistsCount := make(map[restApiV1.ArtistId]int)
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

	albumEntity.ArtistIds = []restApiV1.ArtistId{}

	for artistId, artistCount := range artistsCount {
		if artistCount > len(songIds)/2 {
			albumEntity.ArtistIds = append(albumEntity.ArtistIds, artistId)
		}
	}

	// Reorder artists
	e = s.sortArtistIds(txn, albumEntity.ArtistIds)
	if e != nil {
		return e
	}

	artistIdsChanged := !isArtistIdsEqual(albumOldArtistIds, albumEntity.ArtistIds)
	isUpdatedArtistMetaInAlbumArtistIds := false
	if updateArtistMetaArtistId != nil {
		for _, artistId := range albumEntity.ArtistIds {
			if artistId == *updateArtistMetaArtistId {
				isUpdatedArtistMetaInAlbumArtistIds = true
				break
			}
		}
	}

	if artistIdsChanged || isUpdatedArtistMetaInAlbumArtistIds {
		// Update Song AlbumArtists
		for _, songId := range songIds {
			s.updateSongAlbumArtists(txn, songId, albumEntity.ArtistIds)
		}

		// Update album
		albumEntity.UpdateTs = time.Now().UnixNano()

		e := txn.Update(&albumEntity)
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

func (s *Service) DeleteAlbum(externalTrn storm.Node, albumId restApiV1.AlbumId) (*restApiV1.Album, error) {
	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(true)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	deleteTs := time.Now().UnixNano()

	var albumEntity entity.AlbumEntity
	e = txn.One("Id", albumId, &albumEntity)
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

	// Delete album
	e = txn.DeleteStruct(&albumEntity)
	if e != nil {
		return nil, e
	}

	// Archive albumId
	e = txn.Save(&entity.DeletedAlbumEntity{Id: albumEntity.Id, DeleteTs: deleteTs})
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var album restApiV1.Album
	albumEntity.Fill(&album)

	return &album, nil
}

func (s *Service) GetDeletedAlbumIds(externalTrn storm.Node, fromTs int64) ([]restApiV1.AlbumId, error) {
	var e error

	albumIds := []restApiV1.AlbumId{}
	deletedAlbumEntities := []entity.DeletedAlbumEntity{}

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(false)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	query := txn.Select(q.Gte("DeleteTs", fromTs))

	e = query.Find(&deletedAlbumEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	for _, deletedAlbumEntity := range deletedAlbumEntities {
		albumIds = append(albumIds, deletedAlbumEntity.Id)
	}

	return albumIds, nil
}

func (s *Service) getAlbumIdFromAlbumName(externalTrn storm.Node, albumName string, lastAlbumId *restApiV1.AlbumId) (restApiV1.AlbumId, error) {
	var e error

	var albumId restApiV1.AlbumId

	if albumName != "" {

		// Check available transaction
		txn := externalTrn
		if txn == nil {
			txn, e = s.Db.Begin(true)
			if e != nil {
				return "", e
			}
			defer txn.Rollback()
		}

		var albums []restApiV1.Album
		albums, e = s.ReadAlbums(txn, &restApiV1.AlbumFilter{Name: &albumName})
		if e != nil {
			return "", e
		}
		if len(albums) > 0 {
			// Link the song to an existing album
			if lastAlbumId == nil {
				albumId = albums[0].Id
			} else {
				for _, album := range albums {
					if album.Id == *lastAlbumId {
						albumId = *lastAlbumId
					}
				}
				if albumId == "" {
					// Create the album before linking it to the song
					var album, e = s.CreateAlbum(txn, &restApiV1.AlbumMeta{Name: albumName})
					if e != nil {
						return "", e
					}
					albumId = album.Id
				}
			}
		} else {
			// Create the album before linking it to the song
			var album, e = s.CreateAlbum(txn, &restApiV1.AlbumMeta{Name: albumName})
			if e != nil {
				return "", e
			}
			albumId = album.Id
		}

		// Commit transaction
		if externalTrn == nil {
			txn.Commit()
		}

	}

	return albumId, nil
}
