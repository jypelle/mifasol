package oldstore

import (
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/codec/gob"
	"github.com/jypelle/mifasol/internal/srv/config"
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

	return service
}

func (s *OldStore) Close() error {
	return s.Db.Close()
}
