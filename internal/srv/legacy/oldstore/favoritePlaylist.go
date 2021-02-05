package oldstore

import (
	"github.com/asdine/storm/v3"
	"github.com/jypelle/mifasol/internal/srv/legacy/oldentity"
	"github.com/jypelle/mifasol/internal/srv/storeerror"
	"github.com/jypelle/mifasol/restApiV1"
	"time"
)

func (s *OldStore) ReadFavoritePlaylists(externalTrn storm.Node, filter *restApiV1.FavoritePlaylistFilter) ([]restApiV1.FavoritePlaylist, error) {
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

	favoritePlaylistEntities := []oldentity.FavoritePlaylistEntity{}

	if filter.FromTs != nil {
		e = txn.Range("UpdateTs", *filter.FromTs, time.Now().UnixNano(), &favoritePlaylistEntities)
	} else {
		e = txn.All(&favoritePlaylistEntities)
	}

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	favoritePlaylists := []restApiV1.FavoritePlaylist{}

	for _, favoritePlaylistEntity := range favoritePlaylistEntities {
		var favoritePlaylist restApiV1.FavoritePlaylist
		favoritePlaylistEntity.Fill(&favoritePlaylist)
		favoritePlaylists = append(favoritePlaylists, favoritePlaylist)
	}

	return favoritePlaylists, nil
}

func (s *OldStore) ReadFavoritePlaylist(externalTrn storm.Node, favoritePlaylistId restApiV1.FavoritePlaylistId) (*restApiV1.FavoritePlaylist, error) {
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

	var favoritePlaylistEntity oldentity.FavoritePlaylistEntity
	e = txn.One("Id", string(favoritePlaylistId.UserId)+":"+string(favoritePlaylistId.PlaylistId), &favoritePlaylistEntity)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, storeerror.ErrNotFound
		}
		return nil, e
	}

	var favoritePlaylist restApiV1.FavoritePlaylist
	favoritePlaylistEntity.Fill(&favoritePlaylist)

	return &favoritePlaylist, nil
}
