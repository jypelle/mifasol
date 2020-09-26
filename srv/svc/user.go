package svc

import (
	"github.com/asdine/storm/v3"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/srv/entity"
	"github.com/jypelle/mifasol/tool"
	"time"
)

const DefaultUserName = "mifasol"
const DefaultUserPassword = "mifasol"

func (s *Service) ReadUsers(externalTrn storm.Node, filter *restApiV1.UserFilter) ([]restApiV1.User, error) {
	if s.ServerConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "ReadUsers")
	}
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

	userEntities := []entity.UserEntity{}

	if filter.FromTs != nil {
		e = txn.Range("UpdateTs", *filter.FromTs, time.Now().UnixNano(), &userEntities)
	} else {
		e = txn.All(&userEntities)
	}

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	users := []restApiV1.User{}

	for _, userEntity := range userEntities {
		if filter.AdminFg != nil && *filter.AdminFg != userEntity.AdminFg {
			continue
		}

		var user restApiV1.User
		userEntity.Fill(&user)
		users = append(users, user)
	}

	return users, nil
}

func (s *Service) ReadUser(externalTrn storm.Node, userId restApiV1.UserId) (*restApiV1.User, error) {
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

	var userEntity entity.UserEntity
	e = txn.One("Id", userId, &userEntity)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}

	var user restApiV1.User
	userEntity.Fill(&user)

	return &user, nil
}

func (s *Service) ReadUserEntityByUserName(externalTrn storm.Node, userName string) (*entity.UserEntity, error) {

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

	var userEntity entity.UserEntity
	e = txn.One("Name", userName, &userEntity)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}

	return &userEntity, nil
}

func (s *Service) CreateUser(externalTrn storm.Node, userMetaComplete *restApiV1.UserMetaComplete) (*restApiV1.User, error) {
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

	// Store user
	now := time.Now().UnixNano()

	userEntity := entity.UserEntity{
		Id:         restApiV1.UserId(tool.CreateUlid()),
		CreationTs: now,
		UpdateTs:   now,
		Password:   userMetaComplete.Password,
	}
	userEntity.LoadMeta(&userMetaComplete.UserMeta)

	e = txn.Save(&userEntity)
	if e != nil {
		return nil, e
	}

	// Add incoming playlist to favorite playlist
	favoritePlaylistMeta := &restApiV1.FavoritePlaylistMeta{restApiV1.FavoritePlaylistId{UserId: userEntity.Id, PlaylistId: restApiV1.IncomingPlaylistId}}
	_, e = s.CreateFavoritePlaylist(txn, favoritePlaylistMeta, false)
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var user restApiV1.User
	userEntity.Fill(&user)

	return &user, nil
}

func (s *Service) UpdateUser(externalTrn storm.Node, userId restApiV1.UserId, userMetaComplete *restApiV1.UserMetaComplete) (*restApiV1.User, error) {
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

	var userEntity entity.UserEntity
	e = txn.One("Id", userId, &userEntity)
	if e != nil {
		return nil, e
	}

	userEntity.LoadMeta(&userMetaComplete.UserMeta)

	// Update only non void password
	if userMetaComplete.Password != "" {
		userEntity.Password = userMetaComplete.Password
	}

	userEntity.UpdateTs = time.Now().UnixNano()

	// Update user
	e = txn.Save(&userEntity)
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var user restApiV1.User
	userEntity.Fill(&user)

	return &user, nil
}

func (s *Service) DeleteUser(externalTrn storm.Node, userId restApiV1.UserId) (*restApiV1.User, error) {
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

	var userEntity entity.UserEntity
	e = txn.One("Id", userId, &userEntity)
	if e != nil {
		return nil, e
	}

	// Remove user from owned playlists
	playlistIds, e := s.GetPlaylistIdsFromOwnerUserId(txn, userId)
	if e != nil {
		return nil, e
	}
	for _, playlistId := range playlistIds {
		playList, e := s.ReadPlaylist(txn, playlistId)
		if e != nil {
			return nil, e
		}

		newOwnerUserIds := make([]restApiV1.UserId, 0)
		for _, currentUserId := range playList.OwnerUserIds {
			if currentUserId != userId {
				newOwnerUserIds = append(newOwnerUserIds, currentUserId)
			}
		}
		playList.OwnerUserIds = newOwnerUserIds
		_, e = s.UpdatePlaylist(txn, playlistId, &playList.PlaylistMeta, false)
		if e != nil {
			return nil, e
		}
	}

	// Delete user's favorite playlists
	favoritePlaylistEntities := []entity.FavoritePlaylistEntity{}
	e = txn.Find("UserId", userId, &favoritePlaylistEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}
	for _, favoritePlaylistEntity := range favoritePlaylistEntities {
		s.DeleteFavoritePlaylist(txn, restApiV1.FavoritePlaylistId{UserId: favoritePlaylistEntity.UserId, PlaylistId: favoritePlaylistEntity.PlaylistId})
	}

	// Delete user
	e = txn.DeleteStruct(&userEntity)
	if e != nil {
		return nil, e
	}

	// Archive userId
	e = txn.Save(&entity.DeletedUserEntity{Id: userEntity.Id, DeleteTs: deleteTs})
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var user restApiV1.User
	userEntity.Fill(&user)

	return &user, nil
}

func (s *Service) GetDeletedUserIds(externalTrn storm.Node, fromTs int64) ([]restApiV1.UserId, error) {
	if s.ServerConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "GetDeletedUserIds")
	}
	var e error

	userIds := []restApiV1.UserId{}
	deletedUserEntities := []entity.DeletedUserEntity{}

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(false)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	e = txn.Range("DeleteTs", fromTs, time.Now().UnixNano(), &deletedUserEntities)

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	for _, deletedUserEntity := range deletedUserEntities {
		userIds = append(userIds, deletedUserEntity.Id)
	}

	return userIds, nil
}
