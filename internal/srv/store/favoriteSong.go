package store

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/jypelle/mifasol/internal/srv/entity"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"time"
)

func (s *Store) ReadFavoriteSongs(externalTrn *sqlx.Tx, filter *restApiV1.FavoriteSongFilter) ([]restApiV1.FavoriteSong, error) {
	if s.serverConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "ReadFavoriteSongs")
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

	rows, err := txn.NamedQuery(
		`SELECT
				f.*
			FROM favorite_song f
			WHERE 1>0
			`+tool.TernStr(filter.FromTs != nil, "AND f.update_ts >= :from_ts ", "")+`
			ORDER BY a.update_ts ASC
		`,
		queryArgs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	favoriteSongs := []restApiV1.FavoriteSong{}

	for rows.Next() {
		var favoriteSongEntity entity.FavoriteSongEntity
		err = rows.StructScan(&favoriteSongEntity)
		if err != nil {
			return nil, err
		}

		var favoriteSong restApiV1.FavoriteSong
		favoriteSongEntity.Fill(&favoriteSong)
		favoriteSongs = append(favoriteSongs, favoriteSong)
	}

	return favoriteSongs, nil
}

func (s *Store) CreateFavoriteSong(externalTrn *sqlx.Tx, favoriteSongMeta *restApiV1.FavoriteSongMeta, check bool) (*restApiV1.FavoriteSong, error) {
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

	var favoriteSongEntity entity.FavoriteSongEntity

	err = txn.Get(&favoriteSongEntity, "SELECT * FROM favorite_song WHERE user_id = ? AND song_id = ?", favoriteSongMeta.Id.UserId, favoriteSongMeta.Id.SongId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == sql.ErrNoRows {
		// Store favorite song
		now := time.Now().UnixNano()

		favoriteSongEntity = entity.FavoriteSongEntity{
			UpdateTs: now,
		}
		favoriteSongEntity.LoadMeta(favoriteSongMeta)

		_, err = txn.NamedExec(`
			INSERT INTO	favorite_song (
			    user_id,
				song_id,
			    update_ts
			)
			VALUES (
			    :user_id,
				:song_id,
				:update_ts
			)
		`, &favoriteSongEntity)
		if err != nil {
			return nil, err
		}

		// delete existing deletedFavoriteSong
		queryArgs := make(map[string]interface{})
		queryArgs["user_id"] = favoriteSongMeta.Id.UserId
		queryArgs["song_id"] = favoriteSongMeta.Id.SongId

		_, err = txn.NamedExec(`
			DELETE FROM deleted_favorite_song
			WHERE user_id = :user_id and song_id = :song_id
		`, queryArgs)
		if err != nil {
			return nil, err
		}

		// Force resync on linked favoritePlaylist
		queryArgs = make(map[string]interface{})
		queryArgs["user_id"] = favoriteSongEntity.UserId
		queryArgs["song_id"] = favoriteSongEntity.SongId
		_, err = txn.NamedExec(`
			UPDATE favorite_playlist
			SET update_ts = :update_ts
			WHERE user_id = :user_id AND playlist_id in (
				select distinct playlist_id
				from playlist_song ps
				join favorite_playlist fp
				on fp.playlist_id = ps.playlist_id and fp.user_id = :user_id
				where ps.song_id = :fs.song_id
			)
		`, queryArgs)
		if err != nil {
			return nil, err
		}

		// Commit transaction
		if externalTrn == nil {
			txn.Commit()
		}
	}

	var favoriteSong restApiV1.FavoriteSong
	favoriteSongEntity.Fill(&favoriteSong)

	return &favoriteSong, nil
}

func (s *Store) DeleteFavoriteSong(externalTrn *sqlx.Tx, favoriteSongId restApiV1.FavoriteSongId) (*restApiV1.FavoriteSong, error) {
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

	var favoriteSongEntity entity.FavoriteSongEntity
	err = txn.Get(&favoriteSongEntity, "SELECT * FROM favorite_song WHERE user_id = ? AND song_id = ?", favoriteSongId.UserId, favoriteSongId.SongId)
	if err != nil {
		return nil, err
	}

	// Delete favoriteSong
	_, err = txn.Exec("DELETE FROM favorite_song WHERE user_id = ? AND song_id = ?", favoriteSongId.UserId, favoriteSongId.SongId)
	if err != nil {
		return nil, err
	}

	deleteTs := time.Now().UnixNano()

	// Archive favoriteSongId deletion
	_, err = txn.NamedExec(`
			INSERT INTO	deleted_favorite_song (
			    user_id,
			    song_id,
				delete_ts
			)
			VALUES (
			    :user_id,
			    :song_id,
				:delete_ts
			)
	`, &entity.DeletedFavoriteSongEntity{
		UserId:   favoriteSongEntity.UserId,
		SongId:   favoriteSongEntity.SongId,
		DeleteTs: deleteTs})
	if err != nil {
		return nil, err
	}

	// Force resync on linked favoritePlaylist
	queryArgs := make(map[string]interface{})
	queryArgs["user_id"] = favoriteSongEntity.UserId
	queryArgs["song_id"] = favoriteSongEntity.SongId
	_, err = txn.NamedExec(`
			UPDATE favorite_playlist
			SET update_ts = :update_ts
			WHERE user_id = :user_id AND playlist_id in (
				select distinct playlist_id
				from playlist_song ps
				join favorite_playlist fp
				on fp.playlist_id = ps.playlist_id and fp.user_id = :user_id
				where ps.song_id = :fs.song_id
			)
		`, queryArgs)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var favoriteSong restApiV1.FavoriteSong
	favoriteSongEntity.Fill(&favoriteSong)

	return &favoriteSong, nil
}

func (s *Store) GetDeletedFavoriteSongIds(externalTrn *sqlx.Tx, fromTs int64) ([]restApiV1.FavoriteSongId, error) {
	if s.serverConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "GetDeletedFavoriteSongIds")
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
					d.*
				FROM deleted_favorite_song d
				WHERE d.delete_ts >= :from_ts
				ORDER BY d.delete_ts ASC
			`,
		queryArgs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	favoriteSongIds := []restApiV1.FavoriteSongId{}

	for rows.Next() {
		var deletedFavoriteSongEntity entity.DeletedFavoriteSongEntity
		err = rows.StructScan(&deletedFavoriteSongEntity)
		if err != nil {
			return nil, err
		}

		favoriteSongIds = append(
			favoriteSongIds,
			restApiV1.FavoriteSongId{
				UserId: deletedFavoriteSongEntity.UserId,
				SongId: deletedFavoriteSongEntity.SongId},
		)
	}

	return favoriteSongIds, nil
}

func (s *Store) GetDeletedUserFavoriteSongIds(externalTrn *sqlx.Tx, fromTs int64, userId restApiV1.UserId) ([]restApiV1.SongId, error) {
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
	queryArgs["user_id"] = userId
	queryArgs["from_ts"] = fromTs
	rows, err := txn.NamedQuery(
		`SELECT
					d.*
				FROM deleted_favorite_song d
				WHERE user_id = :user_id AND d.delete_ts >= :from_ts
				ORDER BY d.delete_ts ASC
			`,
		queryArgs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	songIds := []restApiV1.SongId{}

	for rows.Next() {
		var deletedFavoriteSongEntity entity.DeletedFavoriteSongEntity
		err = rows.StructScan(&deletedFavoriteSongEntity)
		if err != nil {
			return nil, err
		}

		songIds = append(songIds, deletedFavoriteSongEntity.SongId)
	}

	return songIds, nil
}
