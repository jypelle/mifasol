package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/jypelle/mifasol/internal/srv/config"
	"github.com/jypelle/mifasol/internal/srv/legacy/oldstore"
	"github.com/jypelle/mifasol/internal/srv/storeerror"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/sirupsen/logrus"
	"os"
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

	// Import old store
	if _, err := os.Stat(serverConfig.GetCompleteConfigOldDbFilename()); err == nil {
		logrus.Info("Import of old database...")
		oldStore := oldstore.NewOldStore(serverConfig)
		err = store.oldStoreImport(oldStore)
		if err != nil {
			logrus.Fatalf("Unable to import old database : %v\n", err)
			os.Remove(serverConfig.GetCompleteConfigDbFilename())
		}
		err := oldStore.Close()
		if err != nil {
			logrus.Fatalf("Unable to close the old database: %v", err)
		}
		err = os.Rename(serverConfig.GetCompleteConfigOldDbFilename(), serverConfig.GetCompleteConfigOldDbFilename()+".old")
		if err != nil {
			logrus.Fatalf("Unable to archive the old database: %v", err)
		}
	}

	// Check existence of the (incoming) playlist
	_, err = store.ReadPlaylist(nil, restApiV1.IncomingPlaylistId)
	if err != nil {
		if err != storeerror.ErrNotFound {
			logrus.Fatalf("Unable to retrieve incoming playlist: %v", err)
		}

		playlistMeta := restApiV1.PlaylistMeta{
			Name:    "(incoming)",
			SongIds: nil,
		}
		_, err = store.CreateInternalPlaylist(nil, restApiV1.IncomingPlaylistId, &playlistMeta, false)
		if err != nil {
			logrus.Fatalf("Unable to create incoming playlist: %v", err)
		}
		logrus.Printf("(incoming) playlist has been created ...")
	}

	// Check existence of at least one admin user
	adminFg := true
	users, e := store.ReadUsers(nil, &restApiV1.UserFilter{AdminFg: &adminFg})
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
		_, e := store.CreateUser(nil, &userMetaComplete)
		if e != nil {
			logrus.Fatalf("Unable to create default mifasol user: %v", e)
		}
		logrus.Printf("No admin user found: the default user/password 'mifasol/mifasol' has been created ...")
	}

	return store
}

func (s *Store) Close() error {
	return s.db.Close()
}
