package oldstore

import (
	"github.com/asdine/storm/v3"
	"github.com/jypelle/mifasol/internal/srv/legacy/oldentity"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"time"
)

func (s *OldStore) ReadArtists(externalTrn storm.Node, filter *restApiV1.ArtistFilter) ([]restApiV1.Artist, error) {
	if s.ServerConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "ReadArtists")
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

	artistEntities := []oldentity.ArtistEntity{}

	if filter.FromTs != nil {
		e = txn.Range("UpdateTs", *filter.FromTs, time.Now().UnixNano(), &artistEntities)
	} else if filter.Name != nil {
		e = txn.Find("Name", *filter.Name, &artistEntities)
	} else {
		e = txn.All(&artistEntities)
	}

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	artists := []restApiV1.Artist{}

	for _, artistEntity := range artistEntities {
		var artist restApiV1.Artist
		artistEntity.Fill(&artist)
		artists = append(artists, artist)
	}

	return artists, nil
}
