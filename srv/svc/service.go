package svc

import (
	"github.com/asdine/storm"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/srv/config"
	"github.com/sirupsen/logrus"
)

type Service struct {
	Db           *storm.DB
	ServerConfig *config.ServerConfig
}

func NewService(db *storm.DB, serverConfig *config.ServerConfig) *Service {

	service := &Service{
		Db:           db,
		ServerConfig: serverConfig,
	}

	// Database upgrade
	e := service.upgrade()
	if e != nil {
		logrus.Fatalf("Unable to upgrade database: %v", e)
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
				Name:    DefaultUserName,
				AdminFg: true,
			},
			Password: DefaultUserPassword,
		}
		_, e := service.CreateUser(nil, &userMetaComplete)
		if e != nil {
			logrus.Fatalf("Unable to create default mifasol user: %v", e)
		}
		logrus.Printf("No admin user found: the default user/password 'mifasol/mifasol' has been created ...")
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

	return service
}
