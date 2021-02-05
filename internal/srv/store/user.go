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

const DefaultUserName = "mifasol"
const DefaultUserPassword = "mifasol"

func (s *Store) ReadUsers(externalTrn *sqlx.Tx, filter *restApiV1.UserFilter) ([]restApiV1.User, error) {
	if s.serverConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "ReadUsers")
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
	if filter.AdminFg != nil {
		queryArgs["admin_fg"] = *filter.AdminFg
	}

	rows, err := txn.NamedQuery(
		`SELECT
				u.*
			FROM user u
			WHERE 1>0
			`+tool.TernStr(filter.FromTs != nil, "AND u.update_ts >= :from_ts ", "")+`
			`+tool.TernStr(filter.AdminFg != nil, "AND u.admin_fg = :admin_fg ", "")+`
			ORDER BY u.name ASC
		`,
		queryArgs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []restApiV1.User{}

	for rows.Next() {
		var userEntity entity.UserEntity
		err = rows.StructScan(&userEntity)
		if err != nil {
			return nil, err
		}

		var user restApiV1.User
		userEntity.Fill(&user)

		users = append(users, user)
	}

	return users, nil
}

func (s *Store) ReadUser(externalTrn *sqlx.Tx, userId restApiV1.UserId) (*restApiV1.User, error) {
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

	var userEntity entity.UserEntity

	err = txn.Get(&userEntity, "SELECT * FROM user WHERE user_id = ?", userId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storeerror.ErrNotFound
		}
		return nil, err
	}

	var user restApiV1.User
	userEntity.Fill(&user)

	return &user, nil
}

func (s *Store) ReadUserByUserName(externalTrn *sqlx.Tx, userName string) (*restApiV1.User, error) {

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

	var userEntity entity.UserEntity

	err = txn.Get(&userEntity, "SELECT * FROM user WHERE name = ?", userName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storeerror.ErrNotFound
		}
		return nil, err
	}

	var user restApiV1.User
	userEntity.Fill(&user)

	return &user, nil
}

func (s *Store) CreateUser(externalTrn *sqlx.Tx, userMetaComplete *restApiV1.UserMetaComplete) (*restApiV1.User, error) {
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

	// Store user
	now := time.Now().UnixNano()

	userEntity := entity.UserEntity{
		UserId:     restApiV1.UserId(tool.CreateUlid()),
		CreationTs: now,
		UpdateTs:   now,
		Password:   userMetaComplete.Password,
	}
	userEntity.LoadMeta(&userMetaComplete.UserMeta)

	_, err = txn.NamedExec(`
			INSERT INTO	user (
				user_id,
				creation_ts,
			    update_ts,
				name,
			    hide_explicit_fg,
			    admin_fg,
			    password
			)
			VALUES (
				:user_id,
				:creation_ts,
			    :update_ts,
				:name,
			    :hide_explicit_fg,
			    :admin_fg,
			    :password
			)
	`, &userEntity)

	// Add incoming playlist to favorite playlist
	favoritePlaylistMeta := &restApiV1.FavoritePlaylistMeta{restApiV1.FavoritePlaylistId{UserId: userEntity.UserId, PlaylistId: restApiV1.IncomingPlaylistId}}
	_, err = s.CreateFavoritePlaylist(txn, favoritePlaylistMeta, false)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var user restApiV1.User
	userEntity.Fill(&user)

	return &user, nil
}

func (s *Store) UpdateUser(externalTrn *sqlx.Tx, userId restApiV1.UserId, userMetaComplete *restApiV1.UserMetaComplete) (*restApiV1.User, error) {
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

	var userEntity entity.UserEntity
	err = txn.Get(&userEntity, "SELECT * FROM user WHERE user_id = ?", userId)
	if err != nil {
		return nil, err
	}

	userEntity.LoadMeta(&userMetaComplete.UserMeta)

	// Update only non void password
	if userMetaComplete.Password != "" {
		userEntity.Password = userMetaComplete.Password
	}

	userEntity.UpdateTs = time.Now().UnixNano()

	// Update user
	_, err = txn.NamedExec(`
		UPDATE user
		SET name = :name,
		    hide_explicit_fg = :hide_explicit_fg,
		    admin_fg = :admin_fg,
		    password = :password,
			update_ts = :update_ts
		WHERE user_id = :user_id
	`, &userEntity)

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var user restApiV1.User
	userEntity.Fill(&user)

	return &user, nil
}

func (s *Store) DeleteUser(externalTrn *sqlx.Tx, userId restApiV1.UserId) (*restApiV1.User, error) {
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

	var userEntity entity.UserEntity
	err = txn.Get(&userEntity, `SELECT * FROM user WHERE user_id = ?`, userId)
	if err != nil {
		return nil, err
	}

	// Remove user from owned playlists
	queryArgs := make(map[string]interface{})
	queryArgs["user_id"] = userId
	queryArgs["update_ts"] = deleteTs
	_, err = txn.NamedExec(`
			UPDATE playlist
			SET update_ts = :update_ts
			WHERE playlist_id in (
			    select playlist_id
			    from playlist_owned_user pou
			    where pou.user_id = :user_id
			)
		`, queryArgs)
	if err != nil {
		return nil, err
	}
	_, err = txn.Exec(`DELETE FROM playlist_owned_user WHERE user_id = ?`, userId)
	if err != nil {
		return nil, err
	}

	// Delete user's favorite playlists
	favoritePlaylistEntities, err := s.ReadFavoritePlaylists(txn, &restApiV1.FavoritePlaylistFilter{UserId: &userId})
	if err != nil {
		return nil, err
	}
	for _, favoritePlaylistEntity := range favoritePlaylistEntities {
		s.DeleteFavoritePlaylist(txn, restApiV1.FavoritePlaylistId{UserId: favoritePlaylistEntity.Id.UserId, PlaylistId: favoritePlaylistEntity.Id.PlaylistId})
	}

	// Delete user's favorite songs
	queryArgs = make(map[string]interface{})
	queryArgs["delete_ts"] = deleteTs
	queryArgs["user_id"] = userId
	_, err = txn.NamedExec(`
			INSERT INTO	deleted_favorite_song (
			    user_id,
			    song_id,
				delete_ts
			)
			SELECT
			    user_id,
			    song_id,
				:delete_ts
			FROM favorite_song
			WHERE user_id = :user_id
	`, queryArgs)
	if err != nil {
		return nil, err
	}

	queryArgs = make(map[string]interface{})
	queryArgs["user_id"] = userId
	_, err = txn.NamedExec(`
			DELETE FROM	favorite_song
			WHERE user_id = :user_id
		`, queryArgs)
	if err != nil {
		return nil, err
	}

	// Delete user
	_, err = txn.Exec(`DELETE FROM user WHERE user_id = ?`, userId)
	if err != nil {
		return nil, err
	}

	// Archive userId
	_, err = txn.NamedExec(`
			INSERT INTO	deleted_user (
			    user_id,
				delete_ts
			)
			VALUES (
			    :user_id,
				:delete_ts
			)
	`, &entity.DeletedUserEntity{UserId: userEntity.UserId, DeleteTs: deleteTs})
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var user restApiV1.User
	userEntity.Fill(&user)

	return &user, nil
}

func (s *Store) GetDeletedUserIds(externalTrn *sqlx.Tx, fromTs int64) ([]restApiV1.UserId, error) {
	if s.serverConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "GetDeletedUserIds")
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
				u.*
			FROM deleted_user u
			WHERE u.delete_ts >= :from_ts
			ORDER BY u.delete_ts ASC
		`,
		queryArgs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userIds := []restApiV1.UserId{}

	for rows.Next() {
		var deletedUserEntity entity.DeletedUserEntity
		err = rows.StructScan(&deletedUserEntity)
		if err != nil {
			return nil, err
		}

		userIds = append(userIds, deletedUserEntity.UserId)
	}

	return userIds, nil
}
