package svc

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/srv/entity"
	"github.com/jypelle/mifasol/tool"
	"sort"
	"time"
)

func (s *Service) ReadArtists(externalTrn storm.Node, filter *restApiV1.ArtistFilter) ([]restApiV1.Artist, error) {
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
	case restApiV1.ArtistOrderByArtistName:
		query = query.OrderBy("Name")
	case restApiV1.ArtistOrderByUpdateTs:
		query = query.OrderBy("UpdateTs")
	default:
	}

	artistEntities := []entity.ArtistEntity{}
	e = query.Find(&artistEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	artists := []restApiV1.Artist{}

	for _, artistEntity := range artistEntities {
		var artist restApiV1.Artist
		artistEntity.Fill(&artist)
		artists = append(artists, artist)
	}

	return artists, nil
}

func (s *Service) ReadArtist(externalTrn storm.Node, artistId restApiV1.ArtistId) (*restApiV1.Artist, error) {
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

	var artistEntity entity.ArtistEntity
	e = txn.One("Id", artistId, &artistEntity)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}

	var artist restApiV1.Artist
	artistEntity.Fill(&artist)

	return &artist, nil
}

func (s *Service) CreateArtist(externalTrn storm.Node, artistMeta *restApiV1.ArtistMeta) (*restApiV1.Artist, error) {
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

	// Store artist
	now := time.Now().UnixNano()

	artistEntity := entity.ArtistEntity{
		Id:         restApiV1.ArtistId(tool.CreateUlid()),
		CreationTs: now,
		UpdateTs:   now,
	}
	artistEntity.LoadMeta(artistMeta)

	e = txn.Save(&artistEntity)
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var artist restApiV1.Artist
	artistEntity.Fill(&artist)

	return &artist, nil

}

func (s *Service) UpdateArtist(externalTrn storm.Node, artistId restApiV1.ArtistId, artistMeta *restApiV1.ArtistMeta) (*restApiV1.Artist, error) {
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

	var artistEntity entity.ArtistEntity
	e = txn.One("Id", artistId, &artistEntity)
	if e != nil {
		return nil, e
	}

	oldName := artistEntity.Name

	artistEntity.LoadMeta(artistMeta)
	artistEntity.UpdateTs = time.Now().UnixNano()

	// Update artist
	e = txn.Update(&artistEntity)
	if e != nil {
		return nil, e
	}

	// Update tags in songs content
	if oldName != artistEntity.Name {
		songIds, e := s.GetSongIdsFromArtistId(txn, artistId)
		if e != nil {
			return nil, e
		}

		for _, songId := range songIds {
			s.UpdateSong(txn, songId, nil, &artistId, false)
		}
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var artist restApiV1.Artist
	artistEntity.Fill(&artist)

	return &artist, nil
}

func (s *Service) DeleteArtist(externalTrn storm.Node, artistId restApiV1.ArtistId) (*restApiV1.Artist, error) {
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

	var artistEntity entity.ArtistEntity
	e = txn.One("Id", artistId, &artistEntity)
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
	e = txn.DeleteStruct(&artistEntity)
	if e != nil {
		return nil, e
	}

	// Archive artistId
	e = txn.Save(&entity.DeletedArtistEntity{Id: artistEntity.Id, DeleteTs: deleteTs})
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var artist restApiV1.Artist
	artistEntity.Fill(&artist)

	return &artist, nil
}

func (s *Service) GetDeletedArtistIds(externalTrn storm.Node, fromTs int64) ([]restApiV1.ArtistId, error) {
	var e error

	artistIds := []restApiV1.ArtistId{}
	deletedArtistEntities := []entity.DeletedArtistEntity{}

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
	e = query.Find(&deletedArtistEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	for _, deletedArtistEntity := range deletedArtistEntities {
		artistIds = append(artistIds, deletedArtistEntity.Id)
	}

	return artistIds, nil
}

func (s *Service) getArtistIdsFromArtistNames(externalTrn storm.Node, artistNames []string) ([]restApiV1.ArtistId, error) {
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

	var artistIds []restApiV1.ArtistId

	for _, artistName := range artistNames {
		artistName = normalizeString(artistName)
		if artistName != "" {
			var artists []restApiV1.Artist
			artists, e = s.ReadArtists(txn, &restApiV1.ArtistFilter{Name: &artistName})

			if e != nil {
				return nil, e
			}
			var artistId restApiV1.ArtistId
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

func isArtistIdsEqual(a, b []restApiV1.ArtistId) bool {
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

func (s *Service) sortArtistIds(txn storm.Node, artistIds []restApiV1.ArtistId) error {

	var artists []*restApiV1.Artist

	for _, artistId := range artistIds {
		artist, e := s.ReadArtist(txn, artistId)
		if e != nil {
			return e
		}
		artists = append(artists, artist)
	}

	sort.Slice(artistIds, func(i, j int) bool {
		artistI := artists[i]
		artistJ := artists[j]
		if artistI.Name < artistJ.Name {
			return true
		}
		if artistI.Name > artistJ.Name {
			return false
		}
		return artistI.Id < artistJ.Id
	})
	return nil
}
