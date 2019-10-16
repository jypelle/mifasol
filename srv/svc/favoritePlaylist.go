package svc

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"mifasol/restApiV1"
	"mifasol/srv/entity"
	"time"
)

func (s *Service) ReadFavoritePlaylists(externalTrn storm.Node, filter *restApiV1.FavoritePlaylistFilter) ([]restApiV1.FavoritePlaylist, error) {
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

	var matchers []q.Matcher

	if filter.FromTs != nil {
		matchers = append(matchers, q.Gte("UpdateTs", *filter.FromTs))
	}

	query := txn.Select(matchers...)

	switch filter.Order {
	case restApiV1.FavoritePlaylistOrderByUpdateTs:
		query = query.OrderBy("UpdateTs")
	default:
	}

	favoritePlaylistEntities := []entity.FavoritePlaylistEntity{}

	e = query.Find(&favoritePlaylistEntities)
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

func (s *Service) CreateFavoritePlaylist(externalTrn storm.Node, favoritePlaylistMeta *restApiV1.FavoritePlaylistMeta, check bool) (*restApiV1.FavoritePlaylist, error) {
	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(true)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	// Store playlist
	now := time.Now().UnixNano()

	favoritePlaylistEntity := entity.FavoritePlaylistEntity{
		UpdateTs: now,
	}
	favoritePlaylistEntity.LoadMeta(favoritePlaylistMeta)

	e = txn.Save(&favoritePlaylistEntity)
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var favoritePlaylist restApiV1.FavoritePlaylist
	favoritePlaylistEntity.Fill(&favoritePlaylist)

	return &favoritePlaylist, nil
}

func (s *Service) DeleteFavoritePlaylist(externalTrn storm.Node, favoritePlaylistId restApiV1.FavoritePlaylistId) (*restApiV1.FavoritePlaylist, error) {
	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(true)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	deleteTs := time.Now().UnixNano()

	var favoritePlaylistEntity entity.FavoritePlaylistEntity
	e = txn.One("Id", favoritePlaylistId, &favoritePlaylistEntity)
	if e != nil {
		return nil, e
	}

	// Archive favoritePlaylistId
	e = txn.Save(&entity.DeletedFavoritePlaylistEntity{Id: favoritePlaylistEntity.Id, DeleteTs: deleteTs})
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var favoritePlaylist restApiV1.FavoritePlaylist
	favoritePlaylistEntity.Fill(&favoritePlaylist)

	return &favoritePlaylist, nil
}

func (s *Service) GetDeletedFavoritePlaylistIds(externalTrn storm.Node, fromTs int64) ([]restApiV1.FavoritePlaylistId, error) {
	var e error

	favoritePlaylistIds := []restApiV1.FavoritePlaylistId{}
	deletedFavoritePlaylistEntities := []entity.DeletedFavoritePlaylistEntity{}

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(false)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	query := txn.Select(q.Gte("DeleteTs", fromTs)).OrderBy("DeleteTs")

	e = query.Find(&deletedFavoritePlaylistEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	for _, deletedFavoritePlaylistEntity := range deletedFavoritePlaylistEntities {
		favoritePlaylistIds = append(favoritePlaylistIds, deletedFavoritePlaylistEntity.Id)
	}

	return favoritePlaylistIds, nil
}
