package oldstore

import (
	"github.com/asdine/storm/v3"
	"github.com/jypelle/mifasol/internal/srv/legacy/oldentity"
	"github.com/jypelle/mifasol/internal/srv/storeerror"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"time"
)

func (s *OldStore) ReadFavoriteSongs(externalTrn storm.Node, filter *restApiV1.FavoriteSongFilter) ([]restApiV1.FavoriteSong, error) {
	if s.ServerConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "ReadFavoriteSongs")
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

	favoriteSongEntities := []oldentity.FavoriteSongEntity{}

	if filter.FromTs != nil {
		e = txn.Range("UpdateTs", *filter.FromTs, time.Now().UnixNano(), &favoriteSongEntities)
	} else {
		e = txn.All(&favoriteSongEntities)
	}

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	favoriteSongs := []restApiV1.FavoriteSong{}

	for _, favoriteSongEntity := range favoriteSongEntities {
		var favoriteSong restApiV1.FavoriteSong
		favoriteSongEntity.Fill(&favoriteSong)
		favoriteSongs = append(favoriteSongs, favoriteSong)
	}

	return favoriteSongs, nil
}

func (s *OldStore) ReadFavoriteSong(externalTrn storm.Node, favoriteSongId restApiV1.FavoriteSongId) (*restApiV1.FavoriteSong, error) {
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

	var favoriteSongEntity oldentity.FavoriteSongEntity
	e = txn.One("Id", string(favoriteSongId.UserId)+":"+string(favoriteSongId.SongId), &favoriteSongEntity)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, storeerror.ErrNotFound
		}
		return nil, e
	}

	var favoriteSong restApiV1.FavoriteSong
	favoriteSongEntity.Fill(&favoriteSong)

	return &favoriteSong, nil
}
