package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/jypelle/mifasol/internal/srv/config"
	"github.com/sirupsen/logrus"
)

type Store struct {
	db           *sqlx.DB
	serverConfig *config.ServerConfig
}

func NewStore(serverConfig *config.ServerConfig) *Store {

	// Open database connection
	db, err := sqlx.Open("sqlite3", serverConfig.GetCompleteConfigDbFilename())
	if err != nil {
		logrus.Fatalf("Unable to connect to the database: %v", err)
	}

	store := &Store{
		db:           db,
		serverConfig: serverConfig,
	}

	// Execute database migration scripts
	if err := store.migrateDatabase(); err != nil {
		logrus.Fatalf("Unable to migrate the database: %v", err)
	}

	// Database upgrade
	/*
		e := service.upgrade()
		if e != nil {
			logrus.Fatalf("Unable to upgrade database: %v", e)
		}

		// Check existence of the (incoming) playlist
		_, e = service.ReadPlaylist(nil, restApiV1.IncomingPlaylistId)
		if e != nil {
			if e != ErrNotFound {
				logrus.Fatalf("Unable to retrieve incoming playlist: %v", e)
			}

			playlistMeta := restApiV1.PlaylistMeta{
				Name:    "(incoming)",
				SongIds: nil,
			}
			_, e = service.CreateInternalPlaylist(nil, restApiV1.IncomingPlaylistId, &playlistMeta, false)
			if e != nil {
				logrus.Fatalf("Unable to create incoming playlist: %v", e)
			}
			logrus.Printf("(incoming) playlist has been created ...")
		}

		// Check existence of at least one admin user
		adminFg := true
		users, e := service.ReadUsers(nil, &restApiV1.UserFilter{AdminFg: &adminFg})
		if e != nil {
			logrus.Fatalf("Unable to retrieve users: %v", e)
		}
		if len(users) == 0 {
			// Create default admin user
			userMetaComplete := restApiV1.UserMetaComplete{
				UserMeta: restApiV1.UserMeta{
					Name:           DefaultUserName,
					HideExplicitFg: false,
					AdminFg:        true,
				},
				Password: DefaultUserPassword,
			}
			_, e := service.CreateUser(nil, &userMetaComplete)
			if e != nil {
				logrus.Fatalf("Unable to create default mifasol user: %v", e)
			}
			logrus.Printf("No admin user found: the default user/password 'mifasol/mifasol' has been created ...")
		}
	*/
	return store
}

func (s *Store) Close() error {
	return s.db.Close()
}
