package svc

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"lyra/restApiV1"
	"lyra/tool"
	"sort"
	"time"
)

func (s *Service) ReadAlbums(externalTrn storm.Node, filter *restApiV1.AlbumFilter) ([]restApiV1.Album, error) {
	albums := []restApiV1.Album{}

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

	err = query.Find(&albums)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	return albums, nil
}

func (s *Service) ReadAlbum(externalTrn storm.Node, albumId string) (*restApiV1.Album, error) {
	var album restApiV1.Album

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

	e := txn.One("Id", albumId, &album)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}

	return &album, nil
}

func (s *Service) CreateAlbum(externalTrn storm.Node, albumMeta *restApiV1.AlbumMeta) (*restApiV1.Album, error) {
	var album *restApiV1.Album

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

	// Store album
	now := time.Now().UnixNano()

	album = &restApiV1.Album{
		Id:         tool.CreateUlid(),
		CreationTs: now,
		UpdateTs:   now,
		AlbumMeta:  *albumMeta,
	}

	e := txn.Save(album)
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return album, nil
}

func (s *Service) UpdateAlbum(externalTrn storm.Node, albumId string, albumMeta *restApiV1.AlbumMeta) (*restApiV1.Album, error) {
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

	album, err := s.ReadAlbum(txn, albumId)
	if err != nil {
		return nil, err
	}

	album.AlbumMeta = *albumMeta

	// Update album
	album.UpdateTs = time.Now().UnixNano()

	e := txn.Update(album)
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

func (s *Service) refreshAlbumArtistIds(externalTrn storm.Node, albumId string, updateArtistMetaArtistId *string) error {
	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(true)
		if err != nil {
			return err
		}
		defer txn.Rollback()
	}

	album, e := s.ReadAlbum(txn, albumId)
	if e != nil {
		return e
	}

	songIds, e := s.GetSongIdsFromAlbumId(txn, albumId)
	if e != nil {
		return e
	}

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

		e := txn.Update(album)
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

func (s *Service) DeleteAlbum(externalTrn storm.Node, albumId string) (*restApiV1.Album, error) {
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

	// Delete album
	e = txn.DeleteStruct(album)
	if e != nil {
		return nil, e
	}

	// Archive albumId
	e = txn.Save(&restApiV1.DeletedAlbum{Id: album.Id, DeleteTs: deleteTs})
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return album, nil
}

func (s *Service) GetDeletedAlbumIds(externalTrn storm.Node, fromTs int64) ([]string, error) {

	albumIds := []string{}
	deletedAlbums := []restApiV1.DeletedAlbum{}

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

	err = query.Find(&deletedAlbums)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	for _, deletedAlbum := range deletedAlbums {
		albumIds = append(albumIds, deletedAlbum.Id)
	}

	return albumIds, nil
}

func (s *Service) getAlbumIdFromAlbumName(externalTrn storm.Node, albumName string, lastAlbumId *string) (string, error) {
	var albumId string

	if albumName != "" {

		// Check available transaction
		txn := externalTrn
		var err error
		if txn == nil {
			txn, err = s.Db.Begin(true)
			if err != nil {
				return "", err
			}
			defer txn.Rollback()
		}

		albums, err := s.ReadAlbums(txn, &restApiV1.AlbumFilter{Name: &albumName})
		if err != nil {
			return "", err
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
					album, err := s.CreateAlbum(txn, &restApiV1.AlbumMeta{Name: albumName})
					if err != nil {
						return "", err
					}
					albumId = album.Id
				}
			}
		} else {
			// Create the album before linking it to the song
			album, err := s.CreateAlbum(txn, &restApiV1.AlbumMeta{Name: albumName})
			if err != nil {
				return "", err
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
