package store

import (
	"database/sql"
	"github.com/jypelle/mifasol/internal/srv/config"
	"github.com/sirupsen/logrus"
)

type Store struct {
	Db           *sql.DB
	ServerConfig *config.ServerConfig
}

func NewStore(db *sql.DB, serverConfig *config.ServerConfig) *Store {

	store := &Store{
		Db:           db,
		ServerConfig: serverConfig,
	}

	// Database upgrade
	if _, err := db.Exec(`
drop table if exists album;
create table album
(
    id text not null primary key,
    creation_ts integer not null,
    update_ts integer not null,
    name text not null
);

drop table if exists artist;
create table artist
(
    id text not null primary key,
    creation_ts integer not null,
    update_ts integer not null,
    name text not null
);

drop table if exists song;
create table song
(
    id text not null primary key,
    creation_ts integer not null,
    update_ts integer not null,
    name text not null,
	format integer not null,
	size integer not null,
	bit_depth integer not null,
	publication_year integer,
	album_id integer not null,
	track_number integer,
	explicit_fg bool not null
);

drop table if exists user;
create table user
(
    id text not null primary key,
    creation_ts integer not null,
    update_ts integer not null,
    name text not null,
	hide_explicit_fg bool not null,
	admin_fg bool not null,
	password text not null
);

`); err != nil {
		logrus.Fatalf("Unable to migrate the database: %v", err)
	}

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
