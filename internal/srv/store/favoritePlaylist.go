package store

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/jypelle/mifasol/internal/srv/entity"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"time"
)

func (s *Store) ReadFavoritePlaylists(externalTrn *sqlx.Tx, filter *restApiV1.FavoritePlaylistFilter) ([]restApiV1.FavoritePlaylist, error) {
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
	if filter.UserId != nil {
		queryArgs["user_id"] = *filter.UserId
	}
	if filter.PlaylistId != nil {
		queryArgs["playlist_id"] = *filter.PlaylistId
	}

	rows, err := txn.NamedQuery(
		`SELECT
				f.*
			FROM favorite_playlist f
			WHERE 1>0
			`+tool.TernStr(filter.FromTs != nil, "AND f.update_ts >= :from_ts ", "")+`
			`+tool.TernStr(filter.UserId != nil, "AND f.user_id = :user_id ", "")+`
			`+tool.TernStr(filter.PlaylistId != nil, "AND f.playlist_id = :playlist_id ", "")+`
			ORDER BY f.update_ts ASC
		`,
		queryArgs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	favoritePlaylists := []restApiV1.FavoritePlaylist{}

	for rows.Next() {
		var favoritePlaylistEntity entity.FavoritePlaylistEntity
		err = rows.StructScan(&favoritePlaylistEntity)
		if err != nil {
			return nil, err
		}

		var favoritePlaylist restApiV1.FavoritePlaylist
		favoritePlaylistEntity.Fill(&favoritePlaylist)
		favoritePlaylists = append(favoritePlaylists, favoritePlaylist)
	}

	return favoritePlaylists, nil
}

func (s *Store) CreateFavoritePlaylist(externalTrn *sqlx.Tx, favoritePlaylistMeta *restApiV1.FavoritePlaylistMeta, check bool) (*restApiV1.FavoritePlaylist, error) {
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

	var favoritePlaylistEntity entity.FavoritePlaylistEntity

	err = txn.Get(&favoritePlaylistEntity, "SELECT * FROM favorite_playlist WHERE user_id = ? AND playlist_id = ?", favoritePlaylistMeta.Id.UserId, favoritePlaylistMeta.Id.PlaylistId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == sql.ErrNoRows {
		// Store favorite playlist
		now := time.Now().UnixNano()

		favoritePlaylistEntity = entity.FavoritePlaylistEntity{
			UpdateTs: now,
		}
		favoritePlaylistEntity.LoadMeta(favoritePlaylistMeta)

		_, err = txn.NamedExec(`
			INSERT INTO	favorite_playlist (
			    user_id,
				playlist_id,
			    update_ts
			)
			VALUES (
			    :user_id,
				:playlist_id,
				:update_ts
			)
		`, &favoritePlaylistEntity)
		if err != nil {
			return nil, err
		}

		// delete existing deletedFavoritePlaylist
		queryArgs := make(map[string]interface{})
		queryArgs["user_id"] = favoritePlaylistMeta.Id.UserId
		queryArgs["playlist_id"] = favoritePlaylistMeta.Id.PlaylistId

		_, err = txn.NamedExec(`
			DELETE FROM deleted_favorite_playlist
			WHERE user_id = :user_id and playlist_id = :playlist_id
		`, queryArgs)
		if err != nil {
			return nil, err
		}
		/*
			// Add favorite playlist songs to favorite songs
			_, err = txn.NamedExec(`
				INSERT INTO	favorite_song (
				    user_id,
					song_id,
				    update_ts
				)
				SELECT DISTINCT
				    :user_id,
					ps.song_id,
					:update_ts
				FROM playlist_song ps
				LEFT JOIN favorite_song fs ON fs.user_id = :user_id AND fs.song_id = ps.song_id
				WHERE ps.playlist_id = :playlist_id
				AND fs.user_id IS NULL

			`, &favoritePlaylistEntity)
			if err != nil {
				return nil, err
			}

			// delete existing deletedFavoriteSong
			queryArgs = make(map[string]interface{})
			queryArgs["user_id"] = favoritePlaylistEntity.UserId
			queryArgs["update_ts"] = favoritePlaylistEntity.UpdateTs
			_, err = txn.NamedExec(`
				DELETE FROM deleted_favorite_song
				where (user_id,song_id) in (select user_id,song_id from favorite_song where user_id = :user_id and update_ts = :update_ts)
			`, queryArgs)
			if err != nil {
				return nil, err
			}

			// force resync on linked favoritePlaylist
			queryArgs = make(map[string]interface{})
			queryArgs["user_id"] = favoritePlaylistEntity.UserId
			queryArgs["update_ts"] = favoritePlaylistEntity.UpdateTs
			_, err = txn.NamedExec(`
				UPDATE favorite_playlist
				SET update_ts = :update_ts
				WHERE user_id = :user_id AND playlist_id in (
				    select distinct playlist_id
				    from favorite_song fs
				    join playlist_song ps using (song_id)
				    join favorite_playlist fp using (playlist_id,user_id)
				    where fs.user_id = :user_id and fs.update_ts = :update_ts
				)
			`, queryArgs)
			if err != nil {
				return nil, err
			}
		*/
		// Commit transaction
		if externalTrn == nil {
			txn.Commit()
		}
	}

	var favoritePlaylist restApiV1.FavoritePlaylist
	favoritePlaylistEntity.Fill(&favoritePlaylist)

	return &favoritePlaylist, nil
}

func (s *Store) DeleteFavoritePlaylist(externalTrn *sqlx.Tx, favoritePlaylistId restApiV1.FavoritePlaylistId) (*restApiV1.FavoritePlaylist, error) {
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

	var favoritePlaylistEntity entity.FavoritePlaylistEntity
	err = txn.Get(&favoritePlaylistEntity, "SELECT * FROM favorite_playlist WHERE user_id = ? AND playlist_id = ?", favoritePlaylistId.UserId, favoritePlaylistId.PlaylistId)
	if err != nil {
		return nil, err
	}

	// Delete favoritePlaylist
	_, err = txn.Exec("DELETE FROM favorite_playlist WHERE user_id = ? AND playlist_id = ?", favoritePlaylistId.UserId, favoritePlaylistId.PlaylistId)
	if err != nil {
		return nil, err
	}

	deleteTs := time.Now().UnixNano()

	// Archive favoritePlaylistId deletion
	_, err = txn.NamedExec(`
			INSERT INTO	deleted_favorite_playlist (
			    user_id,
			    playlist_id,
				delete_ts
			)
			VALUES (
			    :user_id,
			    :playlist_id,
				:delete_ts
			)
	`, &entity.DeletedFavoritePlaylistEntity{
		UserId:     favoritePlaylistEntity.UserId,
		PlaylistId: favoritePlaylistEntity.PlaylistId,
		DeleteTs:   deleteTs})
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var favoritePlaylist restApiV1.FavoritePlaylist
	favoritePlaylistEntity.Fill(&favoritePlaylist)

	return &favoritePlaylist, nil
}

func (s *Store) GetDeletedFavoritePlaylistIds(externalTrn *sqlx.Tx, fromTs int64) ([]restApiV1.FavoritePlaylistId, error) {
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
				FROM deleted_favorite_playlist d
				WHERE d.delete_ts >= :from_ts
				ORDER BY d.delete_ts ASC
			`,
		queryArgs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	favoritePlaylistIds := []restApiV1.FavoritePlaylistId{}

	for rows.Next() {
		var deletedFavoritePlaylistEntity entity.DeletedFavoritePlaylistEntity
		err = rows.StructScan(&deletedFavoritePlaylistEntity)
		if err != nil {
			return nil, err
		}

		favoritePlaylistIds = append(
			favoritePlaylistIds,
			restApiV1.FavoritePlaylistId{
				UserId:     deletedFavoritePlaylistEntity.UserId,
				PlaylistId: deletedFavoritePlaylistEntity.PlaylistId},
		)
	}

	return favoritePlaylistIds, nil
}

func (s *Store) GetDeletedUserFavoritePlaylistIds(externalTrn *sqlx.Tx, fromTs int64, userId restApiV1.UserId) ([]restApiV1.PlaylistId, error) {

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
				FROM deleted_favorite_playlist d
				WHERE user_id = :user_id AND d.delete_ts >= :from_ts
				ORDER BY d.delete_ts ASC
			`,
		queryArgs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	playlistIds := []restApiV1.PlaylistId{}

	for rows.Next() {
		var deletedFavoritePlaylistEntity entity.DeletedFavoritePlaylistEntity
		err = rows.StructScan(&deletedFavoritePlaylistEntity)
		if err != nil {
			return nil, err
		}

		playlistIds = append(playlistIds, deletedFavoritePlaylistEntity.PlaylistId)
	}

	return playlistIds, nil
}
