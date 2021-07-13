package store

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/jypelle/mifasol/internal/srv/entity"
	"github.com/jypelle/mifasol/internal/srv/storeerror"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"sort"
	"time"
)

func (s *Store) ReadArtists(externalTrn *sqlx.Tx, filter *restApiV1.ArtistFilter) ([]restApiV1.Artist, error) {
	if s.serverConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "ReadArtists")
	}

	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	queryArgs := make(map[string]interface{})
	if filter.FromTs != nil {
		queryArgs["from_ts"] = *filter.FromTs
	}
	if filter.Name != nil {
		queryArgs["name"] = *filter.Name
	}
	if filter.SongId != nil {
		queryArgs["song_id"] = *filter.SongId
	}

	orderBy := "a.update_ts ASC"
	if filter.OrderBy != nil {
		if *filter.OrderBy == restApiV1.ArtistFilterOrderByName {
			orderBy = "a.name ASC"
		}
	}

	rows, err := txn.NamedQuery(
		`SELECT
				a.*
			FROM artist a
			`+tool.TernStr(filter.SongId != nil, "JOIN artist_song asg ON asg.artist_id = a.artist_id AND asg.song_id = :song_id ", "")+`
			WHERE 1>0
			`+tool.TernStr(filter.FromTs != nil, "AND a.update_ts >= :from_ts ", "")+`
			`+tool.TernStr(filter.Name != nil, "AND a.name LIKE :name ", "")+`
			ORDER BY `+orderBy,
		queryArgs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	artists := []restApiV1.Artist{}

	for rows.Next() {
		var artistEntity entity.ArtistEntity
		err = rows.StructScan(&artistEntity)
		if err != nil {
			return nil, err
		}

		var artist restApiV1.Artist
		artistEntity.Fill(&artist)

		artists = append(artists, artist)
	}

	return artists, nil

}

func (s *Store) ReadArtist(externalTrn *sqlx.Tx, artistId restApiV1.ArtistId) (*restApiV1.Artist, error) {
	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	var artistEntity entity.ArtistEntity

	err = txn.Get(&artistEntity, "SELECT * FROM artist WHERE artist_id = ?", artistId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storeerror.ErrNotFound
		}
		return nil, err
	}

	var artist restApiV1.Artist
	artistEntity.Fill(&artist)

	return &artist, nil
}

func (s *Store) CreateArtist(externalTrn *sqlx.Tx, artistMeta *restApiV1.ArtistMeta) (*restApiV1.Artist, error) {
	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	// Store artist
	now := time.Now().UnixNano()

	artistEntity := entity.ArtistEntity{
		ArtistId:   restApiV1.ArtistId(tool.CreateUlid()),
		CreationTs: now,
		UpdateTs:   now,
	}
	artistEntity.LoadMeta(artistMeta)

	_, err = txn.NamedExec(`
			INSERT INTO	artist (
			    artist_id,
				creation_ts,
			    update_ts,
				name
			)
			VALUES (
			    :artist_id,
				:creation_ts,
				:update_ts,
				:name
			)
	`, &artistEntity)

	if err != nil {
		return nil, err
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var artist restApiV1.Artist
	artistEntity.Fill(&artist)

	return &artist, nil

}

func (s *Store) UpdateArtist(externalTrn *sqlx.Tx, artistId restApiV1.ArtistId, artistMeta *restApiV1.ArtistMeta) (*restApiV1.Artist, error) {
	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	var artistEntity entity.ArtistEntity
	err = txn.Get(&artistEntity, "SELECT * FROM artist WHERE artist_id = ?", artistId)
	if err != nil {
		return nil, err
	}

	oldName := artistEntity.Name

	artistEntity.LoadMeta(artistMeta)
	artistEntity.UpdateTs = time.Now().UnixNano()

	// Update artist
	_, err = txn.NamedExec(`
		UPDATE artist
		SET name = :name,
			update_ts = :update_ts
		WHERE artist_id = :artist_id
	`, &artistEntity)

	if err != nil {
		return nil, err
	}

	// Update tags in songs content
	if oldName != artistEntity.Name {
		songs, err := s.ReadSongs(txn, &restApiV1.SongFilter{ArtistId: &artistId})
		if err != nil {
			return nil, err
		}
		for _, song := range songs {
			s.UpdateSong(txn, song.Id, nil, &artistId, false)
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

func (s *Store) DeleteArtist(externalTrn *sqlx.Tx, artistId restApiV1.ArtistId) (*restApiV1.Artist, error) {
	if s.serverConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "DeleteArtist")
	}

	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	deleteTs := time.Now().UnixNano()

	var artistEntity entity.ArtistEntity
	err = txn.Get(&artistEntity, `SELECT * FROM artist WHERE artist_id = ?`, artistId)
	if err != nil {
		return nil, err
	}

	// Check songs link
	songs, err := s.ReadSongs(txn, &restApiV1.SongFilter{ArtistId: &artistId})
	if err != nil {
		return nil, err
	}
	if len(songs) > 0 {
		return nil, storeerror.ErrDeleteArtistWithSongs
	}

	// Delete artist
	_, err = txn.Exec("DELETE FROM artist WHERE artist_id = ?", artistId)
	if err != nil {
		return nil, err
	}

	// Archive artistId
	_, err = txn.NamedExec(`
			INSERT INTO	deleted_artist (
			    artist_id,
				delete_ts
			)
			VALUES (
			    :artist_id,
				:delete_ts
			)
	`, &entity.DeletedArtistEntity{ArtistId: artistEntity.ArtistId, DeleteTs: deleteTs})
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var artist restApiV1.Artist
	artistEntity.Fill(&artist)

	return &artist, nil
}

func (s *Store) GetDeletedArtistIds(externalTrn *sqlx.Tx, fromTs int64) ([]restApiV1.ArtistId, error) {
	if s.serverConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "GetDeletedArtistIds")
	}

	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	queryArgs := make(map[string]interface{})
	queryArgs["from_ts"] = fromTs
	rows, err := txn.NamedQuery(
		`SELECT
				a.*
			FROM deleted_artist a
			WHERE a.delete_ts >= :from_ts
			ORDER BY a.delete_ts ASC
		`,
		queryArgs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	artistIds := []restApiV1.ArtistId{}

	for rows.Next() {
		var deletedArtistEntity entity.DeletedArtistEntity
		err = rows.StructScan(&deletedArtistEntity)
		if err != nil {
			return nil, err
		}

		artistIds = append(artistIds, deletedArtistEntity.ArtistId)
	}

	return artistIds, nil
}

func (s *Store) getArtistIdsFromArtistNames(externalTrn *sqlx.Tx, artistNames []string) ([]restApiV1.ArtistId, error) {
	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.db.Beginx()
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

func (s *Store) sortArtistIds(externalTrn *sqlx.Tx, artistIds []restApiV1.ArtistId) error {

	var artists []*restApiV1.Artist

	for _, artistId := range artistIds {
		artist, e := s.ReadArtist(externalTrn, artistId)
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
