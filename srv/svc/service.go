package svc

import (
	"github.com/asdine/storm"
	"mifasol/srv/config"
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

	return service
}
