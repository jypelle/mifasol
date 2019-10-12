package svc

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"lyra/restApiV1"
	"lyra/tool"
	"time"
)

const DefaultUserName = "lyra"
const DefaultUserPassword = "lyra"

func (s *Service) ReadUsers(externalTrn storm.Node, filter *restApiV1.UserFilter) ([]restApiV1.User, error) {
	users := []restApiV1.User{}
	userCompletes := []restApiV1.UserComplete{}

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

	if filter.AdminFg != nil {
		matchers = append(matchers, q.Eq("AdminFg", *filter.AdminFg))
	}

	if filter.Order == restApiV1.UserOrderByUpdateTs {
		matchers = append(matchers, q.Gte("UpdateTs", *filter.FromTs))
	}

	query := txn.Select(matchers...)

	switch filter.Order {
	case restApiV1.UserOrderByUserName:
		query = query.OrderBy("Name")
	case restApiV1.UserOrderByUpdateTs:
		query = query.OrderBy("UpdateTs")
	default:
	}

	err = query.Find(&userCompletes)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	for _, userComplete := range userCompletes {
		users = append(users, userComplete.User)
	}

	return users, nil
}

func (s *Service) ReadUserComplete(externalTrn storm.Node, userId string) (*restApiV1.UserComplete, error) {
	var userComplete restApiV1.UserComplete

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

	e := txn.One("Id", userId, &userComplete)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}

	return &userComplete, nil
}

func (s *Service) ReadUser(externalTrn storm.Node, userId string) (*restApiV1.User, error) {
	userComplete, err := s.ReadUserComplete(externalTrn, userId)
	if err != nil {
		return nil, err
	} else {
		return &userComplete.User, nil
	}
}

func (s *Service) ReadUserCompleteByUserName(externalTrn storm.Node, userName string) (*restApiV1.UserComplete, error) {

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

	var userComplete restApiV1.UserComplete
	e := txn.One("Name", userName, &userComplete)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}

	return &userComplete, nil
}

func (s *Service) CreateUser(externalTrn storm.Node, userMetaComplete *restApiV1.UserMetaComplete) (*restApiV1.User, error) {

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

	// Store user
	now := time.Now().UnixNano()

	userComplete := &restApiV1.UserComplete{
		User: restApiV1.User{
			Id:         tool.CreateUlid(),
			CreationTs: now,
			UpdateTs:   now,
			UserMeta:   userMetaComplete.UserMeta,
		},
		Password: userMetaComplete.Password,
	}

	e := txn.Save(userComplete)
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return &userComplete.User, nil
}

func (s *Service) UpdateUser(externalTrn storm.Node, userId string, userMetaComplete *restApiV1.UserMetaComplete) (*restApiV1.User, error) {
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

	userComplete, err := s.ReadUserComplete(txn, userId)
	if err != nil {
		return nil, err
	}

	userComplete.UserMeta = userMetaComplete.UserMeta

	// Update only non void password
	if userMetaComplete.Password != "" {
		userComplete.Password = userMetaComplete.Password
	}

	// Update user
	userComplete.UpdateTs = time.Now().UnixNano()

	e := txn.Update(userComplete)
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return &userComplete.User, nil
}

func (s *Service) DeleteUser(externalTrn storm.Node, userId string) (*restApiV1.User, error) {
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

	user, e := s.ReadUserComplete(txn, userId)
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

		newOwnerUserIds := make([]string, 0)
		for _, currentUserId := range playList.OwnerUserIds {
			if currentUserId != userId {
				newOwnerUserIds = append(newOwnerUserIds, currentUserId)
			}
		}
		playList.OwnerUserIds = newOwnerUserIds
		_, e = s.UpdatePlaylist(txn, playlistId, &playList.PlaylistMeta)
		if e != nil {
			return nil, e
		}
	}

	// Delete user
	e = txn.DeleteStruct(user)
	if e != nil {
		return nil, e
	}

	// Archive userId
	e = txn.Save(&restApiV1.DeletedUser{Id: user.Id, DeleteTs: deleteTs})
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return &user.User, nil
}

func (s *Service) GetDeletedUserIds(externalTrn storm.Node, fromTs int64) ([]string, error) {
	userIds := []string{}
	deletedUsers := []restApiV1.DeletedUser{}

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

	err = query.Find(&deletedUsers)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	for _, deletedUser := range deletedUsers {
		userIds = append(userIds, deletedUser.Id)
	}

	return userIds, nil
}
