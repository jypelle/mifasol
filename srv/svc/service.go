package svc

import (
	"github.com/dgraph-io/badger"
	"lyra/srv/config"
)

type Service struct {
	Db           *badger.DB
	ServerConfig *config.ServerConfig
}

func NewService(db *badger.DB, serverConfig *config.ServerConfig) *Service {

	service := &Service{
		Db:           db,
		ServerConfig: serverConfig,
	}

	return service
}
