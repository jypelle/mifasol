package oldstore

import (
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/codec/gob"
	"github.com/jypelle/mifasol/internal/srv/config"
	"github.com/jypelle/mifasol/internal/srv/storeerror"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/sirupsen/logrus"
)

type OldStore struct {
	Db           *storm.DB
	ServerConfig *config.ServerConfig
}

func NewOldStore(serverConfig *config.ServerConfig) *OldStore {

	db, err := storm.Open(serverConfig.GetCompleteConfigOldDbFilename(), storm.Codec(gob.Codec))
	if err != nil {
		logrus.Fatalf("Unable to connect to the old database: %v", err)
	}

	service := &OldStore{
		Db:           db,
		ServerConfig: serverConfig,
	}

	// Check existence of the (incoming) playlist
	_, e := service.ReadPlaylist(nil, restApiV1.IncomingPlaylistId)
	if e != nil {
		if e != storeerror.ErrNotFound {
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

	return service
}

func (s *OldStore) Close() error {
	return s.Db.Close()
}
