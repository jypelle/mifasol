package svc

import (
	"encoding/json"
	"github.com/dgraph-io/badger"
	"lyra/restApiV1"
	"lyra/tool"
	"strings"
	"time"
)

const DefaultUserName = "lyra"
const DefaultUserPassword = "lyra"

func (s *Service) ReadUsers(externalTrn *badger.Txn, filter *restApiV1.UserFilter) ([]*restApiV1.User, error) {
	users := []*restApiV1.User{}

	opts := badger.DefaultIteratorOptions
	switch filter.Order {
	case restApiV1.UserOrderByUserName:
		opts.Prefix = []byte(userNameUserIdPrefix)
		opts.PrefetchValues = false
	case restApiV1.UserOrderByUpdateTs:
		opts.Prefix = []byte(userUpdateTsUserIdPrefix)
		opts.PrefetchValues = false
	default:
		opts.Prefix = []byte(userIdPrefix)
	}

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	if filter.Order == restApiV1.UserOrderByUpdateTs {
		it.Seek([]byte(userUpdateTsUserIdPrefix + indexTs(filter.FromTs)))
	} else {
		it.Rewind()
	}

	for ; it.Valid(); it.Next() {
		var user *restApiV1.UserComplete

		switch filter.Order {
		case restApiV1.UserOrderByUserName,
			restApiV1.UserOrderByUpdateTs:
			key := it.Item().KeyCopy(nil)

			userId := strings.Split(string(key), ":")[2]
			var e error
			user, e = s.ReadUserComplete(txn, userId)
			if e != nil {
				return nil, e
			}
		default:
			encodedUser, e := it.Item().ValueCopy(nil)
			if e != nil {
				return nil, e
			}
			e = json.Unmarshal(encodedUser, &user)
			if e != nil {
				return nil, e
			}
		}

		if filter.AdminFg != nil {
			if user.AdminFg != *filter.AdminFg {
				continue
			}
		}

		users = append(users, &user.User)

	}

	return users, nil
}

func (s *Service) ReadUserComplete(externalTrn *badger.Txn, userId string) (*restApiV1.UserComplete, error) {
	var userComplete *restApiV1.UserComplete

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	item, e := txn.Get(getUserIdKey(userId))
	if e != nil {
		if e == badger.ErrKeyNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}
	encodedUser, e := item.ValueCopy(nil)
	if e != nil {
		return nil, e
	}
	e = json.Unmarshal(encodedUser, &userComplete)
	if e != nil {
		return nil, e
	}

	return userComplete, nil
}

func (s *Service) ReadUser(externalTrn *badger.Txn, userId string) (*restApiV1.User, error) {
	userComplete, err := s.ReadUserComplete(externalTrn, userId)
	if err != nil {
		return nil, err
	} else {
		return &userComplete.User, nil
	}
}

func (s *Service) ReadUserCompleteByUserName(externalTrn *badger.Txn, login string) (*restApiV1.UserComplete, error) {

	opts := badger.DefaultIteratorOptions
	opts.Prefix = []byte(userNameUserIdPrefix + indexString(login) + ":")
	opts.PrefetchValues = false

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	for it.Rewind(); it.Valid(); it.Next() {

		key := it.Item().KeyCopy(nil)

		userId := strings.Split(string(key), ":")[2]

		return s.ReadUserComplete(externalTrn, userId)
	}

	return nil, ErrNotFound
}

func (s *Service) CreateUser(externalTrn *badger.Txn, userMetaComplete *restApiV1.UserMetaComplete) (*restApiV1.User, error) {

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
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

	encodedUser, _ := json.Marshal(userComplete)
	e := txn.Set(getUserIdKey(userComplete.Id), encodedUser)
	if e != nil {
		return nil, e
	}
	// Store user login Index
	e = txn.Set(getUserNameUserIdKey(userComplete.Name, userComplete.Id), nil)
	if e != nil {
		return nil, e
	}

	// Store updateTs Index
	e = txn.Set(getUserUpdateTsUserIdKey(userComplete.UpdateTs, userComplete.Id), nil)
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return &userComplete.User, nil
}

func (s *Service) UpdateUser(externalTrn *badger.Txn, userId string, userMetaComplete *restApiV1.UserMetaComplete) (*restApiV1.User, error) {
	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	userComplete, err := s.ReadUserComplete(txn, userId)
	if err != nil {
		return nil, err
	}

	userOldName := userComplete.Name
	userOldUpdateTs := userComplete.UpdateTs
	userComplete.UserMeta = userMetaComplete.UserMeta

	// Update only non void password
	if userMetaComplete.Password != "" {
		userComplete.Password = userMetaComplete.Password
	}

	// Update user
	userComplete.UpdateTs = time.Now().UnixNano()
	encodedUser, _ := json.Marshal(userComplete)
	e := txn.Set(getUserIdKey(userComplete.Id), encodedUser)
	if e != nil {
		return nil, e
	}

	// Update user name Index
	e = txn.Delete(getUserNameUserIdKey(userOldName, userComplete.Id))
	if e != nil {
		return nil, e
	}
	e = txn.Set(getUserNameUserIdKey(userComplete.Name, userComplete.Id), nil)
	if e != nil {
		return nil, e
	}

	// Update updateTs Index
	e = txn.Delete(getUserUpdateTsUserIdKey(userOldUpdateTs, userComplete.Id))
	if e != nil {
		return nil, e
	}

	e = txn.Set(getUserUpdateTsUserIdKey(userComplete.UpdateTs, userComplete.Id), nil)
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return &userComplete.User, nil
}

func (s *Service) DeleteUser(externalTrn *badger.Txn, userId string) (*restApiV1.User, error) {
	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
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

	// Delete user name index
	e = txn.Delete(getUserNameUserIdKey(user.Name, userId))
	if e != nil {
		return nil, e
	}

	// Delete user updateTs index
	e = txn.Delete(getUserUpdateTsUserIdKey(user.UpdateTs, userId))
	if e != nil {
		return nil, e
	}

	// Delete user
	e = txn.Delete(getUserIdKey(userId))
	if e != nil {
		return nil, e
	}

	// Archive userId
	e = txn.Set(getUserDeleteTsUserIdKey(deleteTs, user.Id), nil)
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return &user.User, nil
}

func (s *Service) GetDeletedUserIds(externalTrn *badger.Txn, fromTs int64) ([]string, error) {

	userIds := []string{}

	opts := badger.DefaultIteratorOptions
	opts.Prefix = []byte(userDeleteTsUserIdPrefix)
	opts.PrefetchValues = false

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	for it.Seek([]byte(userDeleteTsUserIdPrefix + indexTs(fromTs))); it.Valid(); it.Next() {

		key := it.Item().KeyCopy(nil)

		userId := strings.Split(string(key), ":")[2]

		userIds = append(userIds, userId)

	}

	return userIds, nil
}
