package oldstore

import (
	"github.com/asdine/storm/v3"
	"github.com/jypelle/mifasol/internal/srv/legacy/oldentity"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"time"
)

func (s *OldStore) ReadAlbums(externalTrn storm.Node, filter *restApiV1.AlbumFilter) ([]restApiV1.Album, error) {
	if s.ServerConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "ReadAlbums")
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

	albumEntities := []oldentity.AlbumEntity{}

	if filter.FromTs != nil {
		e = txn.Range("UpdateTs", *filter.FromTs, time.Now().UnixNano(), &albumEntities)
	} else if filter.Name != nil {
		e = txn.Find("Name", *filter.Name, &albumEntities)
	} else {
		e = txn.All(&albumEntities)
	}

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	albums := []restApiV1.Album{}

	for _, albumEntity := range albumEntities {
		var album restApiV1.Album
		albumEntity.Fill(&album)
		albums = append(albums, album)
	}

	return albums, nil
}
