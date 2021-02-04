package oldstore

import (
	"github.com/asdine/storm/v3"
	"github.com/jypelle/mifasol/internal/srv/legacy/oldentity"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"time"
)

func (s *OldStore) ReadUsers(externalTrn storm.Node, filter *restApiV1.UserFilter) ([]restApiV1.User, error) {
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

	userEntities := []oldentity.UserEntity{}

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
