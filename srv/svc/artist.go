package svc

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"lyra/restApiV1"
	"lyra/tool"
	"time"
)

func (s *Service) ReadArtists(externalTrn storm.Node, filter *restApiV1.ArtistFilter) ([]restApiV1.Artist, error) {
	artists := []restApiV1.Artist{}

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
	case restApiV1.ArtistOrderByArtistName:
		query = query.OrderBy("Name")
	case restApiV1.ArtistOrderByUpdateTs:
		query = query.OrderBy("UpdateTs")
	default:
	}

	err = query.Find(&artists)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	return artists, nil
}

func (s *Service) ReadArtist(externalTrn storm.Node, artistId string) (*restApiV1.Artist, error) {
	var artist restApiV1.Artist

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

	e := txn.One("Id", artistId, &artist)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}

	return &artist, nil
}

func (s *Service) CreateArtist(externalTrn storm.Node, artistMeta *restApiV1.ArtistMeta) (*restApiV1.Artist, error) {
	var artist *restApiV1.Artist

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

	// Create artist
	now := time.Now().UnixNano()

	artist = &restApiV1.Artist{
		Id:         tool.CreateUlid(),
		CreationTs: now,
		UpdateTs:   now,
		ArtistMeta: *artistMeta,
	}

	e := txn.Save(artist)
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return artist, nil
}

func (s *Service) UpdateArtist(externalTrn storm.Node, artistId string, artistMeta *restApiV1.ArtistMeta) (*restApiV1.Artist, error) {
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

	artist, err := s.ReadArtist(txn, artistId)

	if err != nil {
		return nil, err
	}

	artist.ArtistMeta = *artistMeta
	artist.UpdateTs = time.Now().UnixNano()

	// Update artist
	e := txn.Update(artist)
	if e != nil {
		return nil, e
	}

	// Update tags in songs content
	songIds, e := s.GetSongIdsFromArtistId(txn, artistId)
	if e != nil {
		return nil, e
	}

	for _, songId := range songIds {
		s.UpdateSong(txn, songId, nil, &artistId, false)
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return artist, nil
}

func (s *Service) DeleteArtist(externalTrn storm.Node, artistId string) (*restApiV1.Artist, error) {
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

	artist, e := s.ReadArtist(txn, artistId)

	if e != nil {
		return nil, e
	}

	// Check songs link
	songIds, e := s.GetSongIdsFromArtistId(txn, artistId)
	if e != nil {
		return nil, e
	}
	if len(songIds) > 0 {
		return nil, ErrDeleteArtistWithSongs
	}

	// Delete artist
	e = txn.DeleteStruct(artist)
	if e != nil {
		return nil, e
	}

	// Archive artistId
	e = txn.Save(&restApiV1.DeletedArtist{Id: artist.Id, DeleteTs: deleteTs})
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return artist, nil
}

func (s *Service) GetDeletedArtistIds(externalTrn storm.Node, fromTs int64) ([]string, error) {

	artistIds := []string{}
	deletedArtists := []restApiV1.DeletedArtist{}

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

	err = query.Find(&deletedArtists)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	for _, deletedArtist := range deletedArtists {
		artistIds = append(artistIds, deletedArtist.Id)
	}

	return artistIds, nil
}

func (s *Service) getArtistIdsFromArtistNames(externalTrn storm.Node, artistNames []string) ([]string, error) {

	var artistIds []string

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

	for _, artistName := range artistNames {
		artistName = normalizeString(artistName)
		if artistName != "" {
			artists, err := s.ReadArtists(txn, &restApiV1.ArtistFilter{Name: &artistName})

			if err != nil {
				return nil, err
			}
			var artistId string
			if len(artists) > 0 {
				// Link the song to an existing artist
				artistId = artists[0].Id
			} else {
				// Create the artist before linking it to the song
				artist, err := s.CreateArtist(txn, &restApiV1.ArtistMeta{Name: artistName})
				if err != nil {
					return nil, err
				}
				artistId = artist.Id
			}
			artistIds = append(artistIds, artistId)
		}
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return artistIds, nil
}

func isArtistIdsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
