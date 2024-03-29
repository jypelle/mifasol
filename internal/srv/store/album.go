package store

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/jypelle/mifasol/internal/srv/entity"
	"github.com/jypelle/mifasol/internal/srv/storeerror"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"time"
)

func (s *Store) ReadAlbums(externalTrn *sqlx.Tx, filter *restApiV1.AlbumFilter) ([]restApiV1.Album, error) {

	type tmpAlbumEntity struct {
		entity.AlbumEntity
		ArtistId   sql.NullString `db:"artist_id"`
		ArtistName sql.NullString `db:"artist_name"`
	}

	if s.serverConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "ReadAlbums")
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
	} else if filter.Name != nil {
		queryArgs["name"] = *filter.Name
	}

	orderBy := "a.update_ts ASC"
	if filter.OrderBy != nil {
		if *filter.OrderBy == restApiV1.AlbumFilterOrderByName {
			orderBy = "a.name ASC"
		}
	}

	rows, err := txn.NamedQuery(
		`SELECT
				a.album_id,
				a.creation_ts,
				a.update_ts,
				a.name,
				null as artist_id,
				null as artist_name
			FROM album a
			WHERE 1>0
			`+tool.TernStr(filter.FromTs != nil, "AND a.update_ts >= :from_ts ", "")+`
			`+tool.TernStr(filter.Name != nil, "AND a.name LIKE :name ", "")+`
			UNION ALL
			SELECT
				a.album_id,
				a.creation_ts,
				a.update_ts,
				a.name,
				ar.artist_id,
				ar.name as artist_name
			FROM (
				SELECT
					aa.album_id,
					aa.creation_ts,
					aa.update_ts,
					aa.name,
					count(song_id)/2 as album_minimum_song_count_per_artist
				FROM album aa
				LEFT JOIN song ss using(album_id)
				WHERE 1>0
				`+tool.TernStr(filter.FromTs != nil, "AND aa.update_ts >= :from_ts ", "")+`
				`+tool.TernStr(filter.Name != nil, "AND aa.name LIKE :name ", "")+`
				GROUP BY
					aa.album_id,
					aa.creation_ts,
					aa.update_ts,
					aa.name
			) a
			LEFT JOIN song s using(album_id)
			LEFT JOIN artist_song ars USING (song_id)
			LEFT JOIN artist ar ON ar.artist_id = ars.artist_id
			GROUP BY
				a.album_id,
				a.creation_ts,
				a.update_ts,
				a.name,
				ar.artist_id,
				ar.name
			HAVING count(distinct song_id) > album_minimum_song_count_per_artist
			ORDER BY `+orderBy+`, artist_name, ar.artist_id`,
		queryArgs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	albums := []restApiV1.Album{}

	var currentAlbum *restApiV1.Album
	for rows.Next() {
		var albumEntity tmpAlbumEntity
		err = rows.StructScan(&albumEntity)
		if err != nil {
			return nil, err
		}

		if currentAlbum != nil && currentAlbum.Id != albumEntity.AlbumId {
			// Save currentAlbum
			albums = append(albums, *currentAlbum)
		}

		if currentAlbum == nil || currentAlbum.Id != albumEntity.AlbumId {
			// New currentAlbum
			currentAlbum = &restApiV1.Album{}
			albumEntity.Fill(currentAlbum)
		}

		// Add artistId to current album
		if albumEntity.ArtistId.Valid {
			currentAlbum.ArtistIds = append(currentAlbum.ArtistIds, restApiV1.ArtistId(albumEntity.ArtistId.String))
		}

	}
	if currentAlbum != nil {
		// Save currentAlbum
		albums = append(albums, *currentAlbum)
	}

	return albums, nil
}

func (s *Store) ReadAlbum(externalTrn *sqlx.Tx, albumId restApiV1.AlbumId) (*restApiV1.Album, error) {
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

	var albumEntity entity.AlbumEntity

	err = txn.Get(&albumEntity, `SELECT * FROM album WHERE album_id = ?`, albumId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storeerror.ErrNotFound
		}
		return nil, err
	}

	// Retrieve album artists
	artistIds := []restApiV1.ArtistId{}
	err = txn.Select(
		&artistIds,
		`SELECT
			artist_id
		FROM song s
		JOIN artist_song USING (song_id)
		JOIN artist a USING (artist_id)
		WHERE album_id = ?
		GROUP BY artist_id
		HAVING count(*) > (SELECT count(*)/2 FROM song where album_id = ? )
		ORDER BY a.name, a.artist_id`,
		albumId,
		albumId,
	)
	if err != nil {
		return nil, err
	}

	var album restApiV1.Album
	albumEntity.Fill(&album)
	album.ArtistIds = artistIds

	return &album, nil
}

func (s *Store) CreateAlbum(externalTrn *sqlx.Tx, albumMeta *restApiV1.AlbumMeta) (*restApiV1.Album, error) {
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

	// Store album
	now := time.Now().UnixNano()

	albumEntity := entity.AlbumEntity{
		AlbumId:    restApiV1.AlbumId(tool.CreateUlid()),
		CreationTs: now,
		UpdateTs:   now,
	}
	albumEntity.LoadMeta(albumMeta)

	_, err = txn.NamedExec(`
			INSERT INTO	album (
			    album_id,
				creation_ts,
			    update_ts,
				name
			)
			VALUES (
			    :album_id,
				:creation_ts,
				:update_ts,
				:name
			)
	`, &albumEntity)

	if err != nil {
		return nil, err
	}

	var album restApiV1.Album
	albumEntity.Fill(&album)

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return &album, nil
}

func (s *Store) UpdateAlbum(externalTrn *sqlx.Tx, albumId restApiV1.AlbumId, albumMeta *restApiV1.AlbumMeta) (*restApiV1.Album, error) {
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
	var albumEntity entity.AlbumEntity
	err = txn.Get(&albumEntity, `SELECT * FROM album WHERE album_id = ?`, albumId)
	if err != nil {
		return nil, err
	}

	oldName := albumEntity.Name

	albumEntity.LoadMeta(albumMeta)
	albumEntity.UpdateTs = time.Now().UnixNano()

	// Update album
	_, err = txn.NamedExec(`
		UPDATE album
		SET name = :name,
			update_ts = :update_ts
		WHERE album_id = :album_id
	`, &albumEntity)

	if err != nil {
		return nil, err
	}

	// Update tags in songs content
	if oldName != albumEntity.Name {
		songs, err := s.ReadSongs(txn, &restApiV1.SongFilter{AlbumId: &albumId})
		if err != nil {
			return nil, err
		}

		for _, song := range songs {
			s.UpdateSong(txn, song.Id, nil, nil, false)
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

func (s *Store) DeleteAlbum(externalTrn *sqlx.Tx, albumId restApiV1.AlbumId) (*restApiV1.Album, error) {
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

	var albumEntity entity.AlbumEntity
	err = txn.Get(&albumEntity, "SELECT * FROM album WHERE album_id = ?", albumId)
	if err != nil {
		return nil, err
	}

	// Check songs link
	songs, err := s.ReadSongs(txn, &restApiV1.SongFilter{AlbumId: &albumId})
	if err != nil {
		return nil, err
	}
	if len(songs) > 0 {
		return nil, storeerror.ErrDeleteAlbumWithSongs
	}

	// Delete album
	_, err = txn.Exec("DELETE FROM album WHERE album_id = ?", albumId)
	if err != nil {
		return nil, err
	}

	// Archive albumId
	_, err = txn.NamedExec(`
			INSERT INTO	deleted_album (
			    album_id,
				delete_ts
			)
			VALUES (
			    :album_id,
				:delete_ts
			)
	`, &entity.DeletedAlbumEntity{AlbumId: albumEntity.AlbumId, DeleteTs: deleteTs})
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var album restApiV1.Album
	albumEntity.Fill(&album)

	return &album, nil
}

func (s *Store) GetDeletedAlbumIds(externalTrn *sqlx.Tx, fromTs int64) ([]restApiV1.AlbumId, error) {
	if s.serverConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "GetDeletedAlbumIds")
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
			FROM deleted_album a
			WHERE a.delete_ts >= :from_ts
			ORDER BY a.delete_ts
		`,
		queryArgs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	albumIds := []restApiV1.AlbumId{}

	for rows.Next() {
		var deletedAlbumEntity entity.DeletedAlbumEntity
		err = rows.StructScan(&deletedAlbumEntity)
		if err != nil {
			return nil, err
		}

		albumIds = append(albumIds, deletedAlbumEntity.AlbumId)
	}

	return albumIds, nil
}

func (s *Store) getAlbumIdFromAlbumName(externalTrn *sqlx.Tx, albumName string, lastAlbumId restApiV1.AlbumId) (restApiV1.AlbumId, error) {
	var err error

	var albumId = restApiV1.UnknownAlbumId

	if albumName != "" {

		// Check available transaction
		txn := externalTrn
		if txn == nil {
			txn, err = s.db.Beginx()
			if err != nil {
				return restApiV1.UnknownAlbumId, err
			}
			defer txn.Rollback()
		}

		var albums []restApiV1.Album
		albums, err = s.ReadAlbums(txn, &restApiV1.AlbumFilter{Name: &albumName})
		if err != nil {
			return restApiV1.UnknownAlbumId, err
		}
		if len(albums) > 0 {
			// Link the song to an existing album
			if lastAlbumId == restApiV1.UnknownAlbumId {
				albumId = albums[0].Id
			} else {
				for _, album := range albums {
					if album.Id == lastAlbumId {
						albumId = lastAlbumId
					}
				}
				if albumId == restApiV1.UnknownAlbumId {
					// Create the album before linking it to the song
					var album, e = s.CreateAlbum(txn, &restApiV1.AlbumMeta{Name: albumName})
					if e != nil {
						return restApiV1.UnknownAlbumId, e
					}
					albumId = album.Id
				}
			}
		} else {
			// Create the album before linking it to the song
			var album, e = s.CreateAlbum(txn, &restApiV1.AlbumMeta{Name: albumName})
			if e != nil {
				return restApiV1.UnknownAlbumId, e
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
