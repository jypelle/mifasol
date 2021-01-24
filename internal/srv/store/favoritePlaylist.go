package store

import (
	"database/sql"
	"github.com/asdine/storm/v3"
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

	rows, err := txn.NamedQuery(
		`SELECT
				f.*
			FROM favorite_playlist f
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

	err = txn.Get(&favoritePlaylistEntity, "SELECT * FROM favorite_playlist WHERE user_id = ? AND playlist_id = ?", favoritePlaylistMeta.Id.UserId, favoritePlaylistMeta.Id.PlaylistId )
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
		`, &queryArgs)
		if err != nil {
			return nil, err
		}

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
		queryArgs["user_id"] = favoritePlaylistMeta.Id.UserId
		_, err = txn.NamedExec(`
			DELETE FROM deleted_favorite_song
			where (user_id,song_id) in (select user_id,song_id from favorite_song where user_id = :user_id)
		`, &queryArgs)
		if err != nil {
			return nil, err
		}

		// force resync on linked favoritePlaylist
		queryArgs = make(map[string]interface{})
		queryArgs["user_id"] = favoritePlaylistMeta.Id.UserId
		queryArgs["update_ts"] = favoritePlaylistEntity.UpdateTs
		_, err ======= txn.NamedExec(`
			UPDATE favorite_playlist
			SET update_ts = :update_ts
			WHERE user_id = :user_id AND playlist_id in (
			    select playlist_id
			    from favorite_song fs
			    where fs.user_id = :user_id
			    ....
			)
		`, &queryArgs)
		if err != nil {
			return nil, err
		}

		playlistSongEntities := []entity.PlaylistSongEntity{}

		err = txn.Find("PlaylistId", favoritePlaylistMeta.Id.PlaylistId, &playlistSongEntities)
		if err != nil && err != storm.ErrNotFound {
			return nil, err
		}

		for _, playlistSongEntity := range playlistSongEntities {
			_, err = s.CreateFavoriteSong(txn, &restApiV1.FavoriteSongMeta{Id: restApiV1.FavoriteSongId{UserId: favoritePlaylistMeta.Id.UserId, SongId: playlistSongEntity.SongId}}, false)
			if err != nil {
				return nil, err
			}
		}

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
	err = txn.One("Id", string(favoritePlaylistId.UserId)+":"+string(favoritePlaylistId.PlaylistId), &favoritePlaylistEntity)
	if err != nil {
		return nil, err
	}

	// Delete favoritePlaylist
	err = txn.DeleteStruct(&favoritePlaylistEntity)
	if err != nil {
		return nil, err
	}

	// Archive favoritePlaylistId deletion
	err = txn.Save(entity.NewDeletedFavoritePlaylistEntity(favoritePlaylistId))
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

	favoritePlaylistIds := []restApiV1.FavoritePlaylistId{}
	deletedFavoritePlaylistEntities := []entity.DeletedFavoritePlaylistEntity{}

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	err = txn.Range("DeleteTs", fromTs, time.Now().UnixNano(), &deletedFavoritePlaylistEntities)

	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	for _, deletedFavoritePlaylistEntity := range deletedFavoritePlaylistEntities {
		favoritePlaylistIds = append(favoritePlaylistIds, restApiV1.FavoritePlaylistId{UserId: deletedFavoritePlaylistEntity.UserId, PlaylistId: deletedFavoritePlaylistEntity.PlaylistId})
	}

	return favoritePlaylistIds, nil
}

func (s *Store) GetDeletedUserFavoritePlaylistIds(externalTrn *sqlx.Tx, fromTs int64, userId restApiV1.UserId) ([]restApiV1.PlaylistId, error) {
	var err error

	playlistIds := []restApiV1.PlaylistId{}
	deletedFavoritePlaylistEntities := []entity.DeletedFavoritePlaylistEntity{}

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	err = txn.Range("DeleteTs", fromTs, time.Now().UnixNano(), &deletedFavoritePlaylistEntities)

	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	for _, deletedFavoritePlaylistEntity := range deletedFavoritePlaylistEntities {
		if deletedFavoritePlaylistEntity.UserId == userId {
			playlistIds = append(playlistIds, deletedFavoritePlaylistEntity.PlaylistId)
		}
	}

	return playlistIds, nil
}

func (s *Store) updateFavoritePlaylistsContainingSong(txn storm.Node, userId restApiV1.UserId, songId restApiV1.SongId) error {
	now := time.Now().UnixNano()

	favoritePlaylistEntities := []entity.FavoritePlaylistEntity{}

	e := txn.Find("UserId", userId, &favoritePlaylistEntities)
	if e != nil && e != storm.ErrNotFound {
		return e
	}

	for _, favoritePlaylistEntity := range favoritePlaylistEntities {
		var playlistSongEntity entity.PlaylistSongEntity
		e = txn.One("Id", string(favoritePlaylistEntity.PlaylistId)+":"+string(songId), &playlistSongEntity)
		if e != nil && e != storm.ErrNotFound {
			return e
		}
		if e != storm.ErrNotFound {
			favoritePlaylistEntity.UpdateTs = now

			e = txn.Save(&favoritePlaylistEntity)
			if e != nil {
				return e
			}
		}
	}
	return nil
}
