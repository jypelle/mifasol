package svc

import (
	"github.com/asdine/storm"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/srv/entity"
	"github.com/jypelle/mifasol/version"
	"github.com/sirupsen/logrus"
)

var dbVersion_0_2_3 restApiV1.Version = restApiV1.Version{0, 2, 3}

// upgrade database structures and contents accordingly to application current version
func (s *Service) upgrade() error {
	var dbVersion restApiV1.Version

	var dbVersionEntity entity.VersionEntity
	e := s.Db.Get("dbProperties", "version", &dbVersionEntity)
	if e != nil && e != storm.ErrNotFound {
		return e
	}

	if e == storm.ErrNotFound {
		dbVersion = version.AppVersion
		logrus.Printf("Initializing database version to %s", dbVersion.String())

		dbVersionEntity.LoadMeta(&dbVersion)
		e = s.Db.Set("dbProperties", "version", &dbVersionEntity)
		if e != nil {
			return e
		}
	} else {
		dbVersionEntity.Fill(&dbVersion)
		logrus.Printf("Current database version: %s", dbVersion.String())
	}

	// Upgrade database to 0_2_3
	//	if dbVersion.LowerThan(dbVersion_0_2_3) {
	if true {

		// Update songEntity structure
		txn, e := s.Db.Begin(true)
		if e != nil {
			return e
		}
		songEntities := []entity.SongEntity{}
		e = txn.All(&songEntities)
		if e != nil && e != storm.ErrNotFound {
			return e
		}
		for _, songEntity := range songEntities {
			txn.DeleteStruct(&songEntity)
			if e != nil {
				return e
			}
			txn.Update(&songEntity)
			if e != nil {
				return e
			}
		}

		txn.Commit()

		dbVersion = dbVersion_0_2_3
		logrus.Printf("Upgrading database version to %s", dbVersion.String())

		// Update db version
		dbVersionEntity.LoadMeta(&dbVersion)
		e = s.Db.Set("dbProperties", "version", &dbVersionEntity)
		if e != nil {
			return e
		}
	}

	return nil
}
