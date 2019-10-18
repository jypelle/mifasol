package svc

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/srv/entity"
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

func (s *Service) ReadFavoritePlaylist(externalTrn storm.Node, favoritePlaylistId restApiV1.FavoritePlaylistId) (*restApiV1.FavoritePlaylist, error) {
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

	var favoritePlaylistEntity entity.FavoritePlaylistEntity
	e = txn.One("Id", favoritePlaylistId, &favoritePlaylistEntity)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}

	var favoritePlaylist restApiV1.FavoritePlaylist
	favoritePlaylistEntity.Fill(&favoritePlaylist)

	return &favoritePlaylist, nil
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

	var favoritePlaylistEntity entity.FavoritePlaylistEntity

	e = txn.One("Id", favoritePlaylistMeta.Id, &favoritePlaylistEntity)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}
	if e == storm.ErrNotFound {
		// Store favorite playlist
		now := time.Now().UnixNano()

		favoritePlaylistEntity = entity.FavoritePlaylistEntity{
			UpdateTs: now,
		}
		favoritePlaylistEntity.LoadMeta(favoritePlaylistMeta)

		e = txn.Save(&favoritePlaylistEntity)
		if e != nil {
			return nil, e
		}

		// if previously deletedFavoritePlaylist exists
		var deletedFavoritePlaylistEntity entity.DeletedFavoritePlaylistEntity
		e = txn.One("Id", favoritePlaylistMeta.Id, &deletedFavoritePlaylistEntity)
		if e != nil && e != storm.ErrNotFound {
			return nil, e
		}

		if e == nil {
			// Delete deletedFavoritePlaylist
			e = txn.DeleteStruct(&deletedFavoritePlaylistEntity)
			if e != nil {
				return nil, e
			}
		}

		// Commit transaction
		if externalTrn == nil {
			txn.Commit()
		}
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

	var favoritePlaylistEntity entity.FavoritePlaylistEntity
	e = txn.One("Id", favoritePlaylistId, &favoritePlaylistEntity)
	if e != nil {
		return nil, e
	}

	// Delete favoritePlaylist
	e = txn.DeleteStruct(&favoritePlaylistEntity)
	if e != nil {
		return nil, e
	}

	// Archive favoritePlaylistId
	e = txn.Save(entity.NewDeletedFavoritePlaylistEntity(favoritePlaylistEntity.Id))
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

func (s *Service) GetDeletedUserFavoritePlaylistIds(externalTrn storm.Node, fromTs int64, userId string) ([]string, error) {
	var e error

	playlistIds := []string{}
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

	query := txn.Select(q.Gte("DeleteTs", fromTs), q.Eq("UserId", userId)).OrderBy("DeleteTs")

	e = query.Find(&deletedFavoritePlaylistEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	for _, deletedFavoritePlaylistEntity := range deletedFavoritePlaylistEntities {
		playlistIds = append(playlistIds, deletedFavoritePlaylistEntity.Id.PlaylistId)
	}

	return playlistIds, nil
}
