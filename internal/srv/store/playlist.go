package store

import (
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/jypelle/mifasol/internal/srv/entity"
	"github.com/jypelle/mifasol/internal/srv/storeerror"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"reflect"
	"sort"
	"time"
)

func (s *Store) ReadPlaylists(externalTrn *sqlx.Tx, filter *restApiV1.PlaylistFilter) ([]restApiV1.Playlist, error) {
	if s.serverConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "ReadPlaylists")
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
	if filter.FavoriteUserId != nil {
		queryArgs["favorite_user_id"] = *filter.FavoriteUserId
	}
	if filter.FavoriteFromTs != nil {
		queryArgs["favorite_from_ts"] = *filter.FavoriteFromTs
	}

	rows, err := txn.NamedQuery(
		`SELECT
				p.*
			FROM playlist p
			`+tool.TernStr(filter.FavoriteUserId != nil, "JOIN favorite_playlist fp ON fp.playlist_id = p.playlist_id AND fp.user_id = :favorite_user_id ", "")+`
			WHERE 1>0
			`+tool.TernStr(filter.FromTs != nil, "AND p.update_ts >= :from_ts ", "")+`
			`+tool.TernStr(filter.FavoriteUserId != nil && filter.FavoriteFromTs != nil, "AND (fp.update_ts >= :favorite_from_ts OR p.content_update_ts >= :favorite_from_ts) ", "")+`
			ORDER BY p.update_ts ASC
		`,
		queryArgs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	playlists := []restApiV1.Playlist{}

	for rows.Next() {
		var playlistEntity entity.PlaylistEntity
		err = rows.StructScan(&playlistEntity)
		if err != nil {
			return nil, err
		}

		// TODO: Need optimizations!

		// Retrieve owned users
		playlistOwnedUserEntities := []entity.PlaylistOwnedUserEntity{}
		err = txn.Select(&playlistOwnedUserEntities, "SELECT * FROM playlist_owned_user WHERE playlist_id = ? ORDER BY user_id", playlistEntity.PlaylistId)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, storeerror.ErrNotFound
			}
			return nil, err
		}

		// Retrieve songs
		playlistSongEntities := []entity.PlaylistSongEntity{}
		err = txn.Select(&playlistSongEntities, "SELECT * FROM playlist_song WHERE playlist_id = ? ORDER BY position", playlistEntity.PlaylistId)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, storeerror.ErrNotFound
			}
			return nil, err
		}

		var playlist restApiV1.Playlist
		playlistEntity.Fill(&playlist)
		for _, playlistOwnedUserEntity := range playlistOwnedUserEntities {
			playlist.OwnerUserIds = append(playlist.OwnerUserIds, playlistOwnedUserEntity.UserId)
		}
		for _, playlistSongEntity := range playlistSongEntities {
			playlist.SongIds = append(playlist.SongIds, playlistSongEntity.SongId)
		}

		playlists = append(playlists, playlist)
	}

	return playlists, nil
}

func (s *Store) ReadPlaylist(externalTrn *sqlx.Tx, playlistId restApiV1.PlaylistId) (*restApiV1.Playlist, error) {
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

	var playlistEntity entity.PlaylistEntity

	err = txn.Get(&playlistEntity, "SELECT * FROM playlist WHERE playlist_id = ?", playlistId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storeerror.ErrNotFound
		}
		return nil, err
	}

	// Retrieve owned users
	playlistOwnedUserEntities := []entity.PlaylistOwnedUserEntity{}
	err = txn.Select(&playlistOwnedUserEntities, "SELECT * FROM playlist_owned_user WHERE playlist_id = ? ORDER BY user_id", playlistId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storeerror.ErrNotFound
		}
		return nil, err
	}

	// Retrieve songs
	playlistSongEntities := []entity.PlaylistSongEntity{}
	err = txn.Select(&playlistSongEntities, "SELECT * FROM playlist_song WHERE playlist_id = ? ORDER BY position", playlistId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storeerror.ErrNotFound
		}
		return nil, err
	}

	var playlist restApiV1.Playlist
	playlistEntity.Fill(&playlist)
	for _, playlistOwnedUserEntity := range playlistOwnedUserEntities {
		playlist.OwnerUserIds = append(playlist.OwnerUserIds, playlistOwnedUserEntity.UserId)
	}
	for _, playlistSongEntity := range playlistSongEntities {
		playlist.SongIds = append(playlist.SongIds, playlistSongEntity.SongId)
	}

	return &playlist, nil
}

func (s *Store) GetDeletedPlaylistIds(externalTrn *sqlx.Tx, fromTs int64) ([]restApiV1.PlaylistId, error) {
	if s.serverConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "GetDeletedPlaylistIds")
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
			FROM deleted_playlist d
			WHERE d.delete_ts >= :from_ts
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
		var deletedPlaylistEntity entity.DeletedPlaylistEntity
		err = rows.StructScan(&deletedPlaylistEntity)
		if err != nil {
			return nil, err
		}

		playlistIds = append(playlistIds, deletedPlaylistEntity.PlaylistId)
	}

	return playlistIds, nil
}

func (s *Store) CreatePlaylist(externalTrn *sqlx.Tx, playlistMeta *restApiV1.PlaylistMeta, check bool) (*restApiV1.Playlist, error) {
	return s.CreateInternalPlaylist(externalTrn, "", playlistMeta, check)
}

func (s *Store) CreateInternalPlaylist(externalTrn *sqlx.Tx, playlistId restApiV1.PlaylistId, playlistMeta *restApiV1.PlaylistMeta, check bool) (*restApiV1.Playlist, error) {
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

	// Store playlist
	now := time.Now().UnixNano()

	if playlistId == "" {
		playlistId = restApiV1.PlaylistId(tool.CreateUlid())
	}

	playlistEntity := entity.PlaylistEntity{
		PlaylistId:      playlistId,
		CreationTs:      now,
		UpdateTs:        now,
		ContentUpdateTs: now,
	}
	playlistEntity.LoadMeta(playlistMeta)

	_, err = txn.NamedExec(`
			INSERT INTO	playlist (
			    playlist_id,
				creation_ts,
			    update_ts,
			    content_update_ts,
				name
			)
			VALUES (
			    :playlist_id,
				:creation_ts,
			    :update_ts,
			    :content_update_ts,
				:name
			)`,
		&playlistEntity,
	)
	if err != nil {
		return nil, err
	}

	// Clean owner list
	playlistMeta.OwnerUserIds = tool.DeduplicateUserId(playlistMeta.OwnerUserIds)
	sort.Slice(playlistMeta.OwnerUserIds, func(i, j int) bool {
		return playlistMeta.OwnerUserIds[i] < playlistMeta.OwnerUserIds[j]
	})

	// Create owners link
	for _, ownerUserId := range playlistMeta.OwnerUserIds {

		// Check owner user id
		if check {
			var userEntity entity.UserEntity
			err = txn.Get(&userEntity, `SELECT * FROM user WHERE user_id = ?`, ownerUserId)
			if err != nil {
				return nil, err
			}
		}

		// Store playlist owner
		queryArgs := make(map[string]interface{})
		queryArgs["playlist_id"] = playlistId
		queryArgs["user_id"] = ownerUserId
		_, err = txn.NamedExec(`
				INSERT INTO	playlist_owned_user (
					playlist_id,
					user_id
				)
				VALUES (
					:playlist_id,
					:user_id
				)
				`, queryArgs)
		if err != nil {
			return nil, err
		}

		// Add playlist to owner favorite playlist
		favoritePlaylistMeta := &restApiV1.FavoritePlaylistMeta{restApiV1.FavoritePlaylistId{UserId: ownerUserId, PlaylistId: playlistId}}
		_, err = s.CreateFavoritePlaylist(txn, favoritePlaylistMeta, false)
		if err != nil {
			return nil, err
		}

	}

	// Create songs link
	for position, songId := range playlistMeta.SongIds {
		// Check song id
		if check {
			var songEntity entity.SongEntity
			err = txn.Get(&songEntity, `SELECT * FROM song WHERE song_id = ?`, songId)
			if err != nil {
				return nil, err
			}
		}

		// Store song link
		queryArgs := make(map[string]interface{})
		queryArgs["playlist_id"] = playlistId
		queryArgs["position"] = position
		queryArgs["song_id"] = songId
		_, err = txn.NamedExec(`
				INSERT INTO	playlist_song (
					playlist_id,
					position,
					song_id
				)
				VALUES (
					:playlist_id,
					:position,
					:song_id
				)
				`, queryArgs)
		if err != nil {
			return nil, err
		}
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var playlist restApiV1.Playlist
	playlistEntity.Fill(&playlist)

	return &playlist, nil
}

func (s *Store) UpdatePlaylist(externalTrn *sqlx.Tx, playlistId restApiV1.PlaylistId, playlistMeta *restApiV1.PlaylistMeta, check bool) (*restApiV1.Playlist, error) {
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

	now := time.Now().UnixNano()

	var playlistEntity entity.PlaylistEntity

	err = txn.Get(&playlistEntity, "SELECT * FROM playlist WHERE playlist_id = ?", playlistId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storeerror.ErrNotFound
		}
		return nil, err
	}

	// Retrieve old songs
	playlistOldSongIds := []restApiV1.SongId{}
	err = txn.Select(&playlistOldSongIds, "SELECT song_id FROM playlist_song WHERE playlist_id = ? ORDER BY position", playlistId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storeerror.ErrNotFound
		}
		return nil, err
	}

	playlistOldName := playlistEntity.Name

	playlistEntity.LoadMeta(playlistMeta)

	// Clean owner list
	playlistMeta.OwnerUserIds = tool.DeduplicateUserId(playlistMeta.OwnerUserIds)
	sort.Slice(playlistMeta.OwnerUserIds, func(i, j int) bool {
		return playlistMeta.OwnerUserIds[i] < playlistMeta.OwnerUserIds[j]
	})

	// Detect song list update
	songIdsUpdated := !reflect.DeepEqual(playlistOldSongIds, playlistMeta.SongIds)

	// Update playlist update timestamp
	playlistEntity.UpdateTs = now

	// Update playlist update content timestamp
	if playlistOldName != playlistEntity.Name || songIdsUpdated {
		playlistEntity.ContentUpdateTs = now
	}

	_, err = txn.NamedExec(`
		UPDATE playlist
		SET name = :name,
			update_ts = :update_ts,
			content_update_ts = :content_update_ts
		WHERE playlist_id = :playlist_id
	`, &playlistEntity)

	if err != nil {
		return nil, err
	}

	// Update owner list
	_, err = txn.Exec("DELETE FROM playlist_owned_user WHERE playlist_id = ?", playlistId)
	if err != nil {
		return nil, err
	}

	for _, ownerUserId := range playlistMeta.OwnerUserIds {

		// Check owner user id
		if check {
			var userEntity entity.UserEntity
			err = txn.Get(&userEntity, `SELECT * FROM user WHERE user_id = ?`, ownerUserId)
			if err != nil {
				return nil, err
			}
		}

		// Store playlist owner
		queryArgs := make(map[string]interface{})
		queryArgs["playlist_id"] = playlistId
		queryArgs["user_id"] = ownerUserId
		_, err = txn.NamedExec(`
				INSERT INTO	playlist_owned_user (
					playlist_id,
					user_id
				)
				VALUES (
					:playlist_id,
					:user_id
				)
				`, queryArgs)
		if err != nil {
			return nil, err
		}
	}

	// Update songs list
	if songIdsUpdated {
		_, err = txn.Exec("DELETE FROM playlist_song WHERE playlist_id = ?", playlistId)
		if err != nil {
			return nil, err
		}

		for position, songId := range playlistMeta.SongIds {
			// Check song id
			if check {
				var songEntity entity.SongEntity
				err = txn.Get(&songEntity, `SELECT * FROM song WHERE song_id = ?`, songId)
				if err != nil {
					return nil, err
				}
			}

			// Store song link
			queryArgs := make(map[string]interface{})
			queryArgs["playlist_id"] = playlistId
			queryArgs["position"] = position
			queryArgs["song_id"] = songId
			_, err = txn.NamedExec(`
				INSERT INTO	playlist_song (
					playlist_id,
					position,
					song_id
				)
				VALUES (
					:playlist_id,
					:position,
					:song_id
				)
				`, queryArgs)
			if err != nil {
				return nil, err
			}
		}
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var playlist restApiV1.Playlist
	playlistEntity.Fill(&playlist)

	return &playlist, nil
}

func (s *Store) AddSongToPlaylist(externalTrn *sqlx.Tx, playlistId restApiV1.PlaylistId, songId restApiV1.SongId, check bool) (*restApiV1.Playlist, error) {
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

	now := time.Now().UnixNano()

	var playlistEntity entity.PlaylistEntity
	err = txn.Get(&playlistEntity, `SELECT * FROM playlist WHERE playlist_id = ?`, playlistId)
	if err != nil {
		return nil, err
	}

	// Check song id
	if check {
		var songEntity entity.SongEntity
		err = txn.Get(&songEntity, `SELECT * FROM song WHERE song_id = ?`, songId)
		if err != nil {
			return nil, err
		}
	}

	// Store song link
	queryArgs := make(map[string]interface{})
	queryArgs["playlist_id"] = playlistId
	queryArgs["song_id"] = songId
	_, err = txn.NamedExec(`
			INSERT INTO	playlist_song (
			    playlist_id,
				position,
			    song_id
			)
			SELECT
			    :playlist_id,
				COALESCE(MAX(position)+1,0) as position,
				:song_id
			FROM playlist_song WHERE playlist_id = :playlist_id
	`, queryArgs)
	if err != nil {
		return nil, err
	}

	// Remove 100th song on incoming playlist
	if playlistId == restApiV1.IncomingPlaylistId {
		_, err = txn.NamedExec(`
			UPDATE playlist_song
			SET position = position -1
			WHERE playlist_id = :playlist_id
			AND (SELECT count(*) FROM playlist_song ps2 WHERE ps2.playlist_id = :playlist_id) > 100
		`, queryArgs)
		if err != nil {
			return nil, err
		}

		_, err = txn.NamedExec(`
			DELETE FROM	playlist_song
			WHERE playlist_id = :playlist_id
			AND position < 0
		`, queryArgs)
		if err != nil {
			return nil, err
		}
	}

	// Update playlist update timestamp
	playlistEntity.UpdateTs = now
	playlistEntity.ContentUpdateTs = now
	_, err = txn.NamedExec(`
		UPDATE playlist
		SET update_ts = :update_ts,
			content_update_ts = :content_update_ts
		WHERE playlist_id = :playlist_id
	`, &playlistEntity)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var playlist restApiV1.Playlist
	playlistEntity.Fill(&playlist)

	return &playlist, nil
}

func (s *Store) DeletePlaylist(externalTrn *sqlx.Tx, playlistId restApiV1.PlaylistId) (*restApiV1.Playlist, error) {
	var err error

	// Incoming playlist can't be deleted
	if playlistId == restApiV1.IncomingPlaylistId {
		return nil, errors.New("Incoming playlist can't be deleted")
	}

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

	var playlistEntity entity.PlaylistEntity
	err = txn.Get(&playlistEntity, `SELECT * FROM playlist WHERE playlist_id = ?`, playlistId)
	if err != nil {
		return nil, err
	}

	// Delete favorite playlist link
	favoritePlaylistEntities, err := s.ReadFavoritePlaylists(txn, &restApiV1.FavoritePlaylistFilter{PlaylistId: &playlistId})
	if err != nil {
		return nil, err
	}
	for _, favoritePlaylistEntity := range favoritePlaylistEntities {
		s.DeleteFavoritePlaylist(txn, restApiV1.FavoritePlaylistId{UserId: favoritePlaylistEntity.Id.UserId, PlaylistId: favoritePlaylistEntity.Id.PlaylistId})
	}

	// Delete owners link
	_, err = txn.Exec("DELETE FROM playlist_owned_user WHERE playlist_id = ?", playlistId)
	if err != nil {
		return nil, err
	}

	// Delete songs link
	_, err = txn.Exec("DELETE FROM playlist_song WHERE playlist_id = ?", playlistId)
	if err != nil {
		return nil, err
	}

	// Delete playlist
	_, err = txn.Exec("DELETE FROM playlist WHERE playlist_id = ?", playlistId)
	if err != nil {
		return nil, err
	}

	// Archive playlistId
	_, err = txn.NamedExec(`
			INSERT INTO	deleted_playlist (
			    playlist_id,
				delete_ts
			)
			VALUES (
			    :playlist_id,
				:delete_ts
			)
	`, &entity.DeletedPlaylistEntity{PlaylistId: playlistEntity.PlaylistId, DeleteTs: deleteTs})
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var playlist restApiV1.Playlist
	playlistEntity.Fill(&playlist)

	return &playlist, nil
}
