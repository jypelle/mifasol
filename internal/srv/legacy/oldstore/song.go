package oldstore

import (
	"github.com/asdine/storm/v3"
	"github.com/jypelle/mifasol/internal/srv/legacy/oldentity"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"time"
)

func (s *OldStore) ReadSongs(externalTrn storm.Node) ([]restApiV1.Song, error) {
	if s.ServerConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "ReadSongs")
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

	songEntities := []oldentity.SongEntity{}

	e = txn.All(&songEntities)

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	songs := []restApiV1.Song{}

	for _, songEntity := range songEntities {
		var song restApiV1.Song
		songEntity.Fill(&song)
		songs = append(songs, song)
	}

	return songs, nil
}
