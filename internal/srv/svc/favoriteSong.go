package svc

import (
	"github.com/asdine/storm/v3"
	"github.com/jypelle/mifasol/internal/srv/entity"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"time"
)

func (s *Service) ReadFavoriteSongs(externalTrn storm.Node, filter *restApiV1.FavoriteSongFilter) ([]restApiV1.FavoriteSong, error) {
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

	favoriteSongEntities := []entity.FavoriteSongEntity{}

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

func (s *Service) ReadFavoriteSong(externalTrn storm.Node, favoriteSongId restApiV1.FavoriteSongId) (*restApiV1.FavoriteSong, error) {
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

	var favoriteSongEntity entity.FavoriteSongEntity
	e = txn.One("Id", string(favoriteSongId.UserId)+":"+string(favoriteSongId.SongId), &favoriteSongEntity)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}

	var favoriteSong restApiV1.FavoriteSong
	favoriteSongEntity.Fill(&favoriteSong)

	return &favoriteSong, nil
}

func (s *Service) CreateFavoriteSong(externalTrn storm.Node, favoriteSongMeta *restApiV1.FavoriteSongMeta, check bool) (*restApiV1.FavoriteSong, error) {
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

	var favoriteSongEntity entity.FavoriteSongEntity

	e = txn.One("Id", string(favoriteSongMeta.Id.UserId)+":"+string(favoriteSongMeta.Id.SongId), &favoriteSongEntity)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}
	if e == storm.ErrNotFound {
		// Store favorite song
		now := time.Now().UnixNano()

		favoriteSongEntity = entity.FavoriteSongEntity{
			UpdateTs: now,
		}
		favoriteSongEntity.LoadMeta(favoriteSongMeta)

		e = txn.Save(&favoriteSongEntity)
		if e != nil {
			return nil, e
		}

		// if previously deletedFavoriteSong exists
		var deletedFavoriteSongEntity entity.DeletedFavoriteSongEntity
		e = txn.One("Id", string(favoriteSongMeta.Id.UserId)+":"+string(favoriteSongMeta.Id.SongId), &deletedFavoriteSongEntity)
		if e != nil && e != storm.ErrNotFound {
			return nil, e
		}

		if e == nil {
			// Delete deletedFavoriteSong
			e = txn.DeleteStruct(&deletedFavoriteSongEntity)
			if e != nil {
				return nil, e
			}
		}

		// Force resync on linked favoritePlaylist
		e = s.updateFavoritePlaylistsContainingSong(txn, favoriteSongMeta.Id.UserId, favoriteSongMeta.Id.SongId)
		if e != nil {
			return nil, e
		}

		// Commit transaction
		if externalTrn == nil {
			txn.Commit()
		}
	}

	var favoriteSong restApiV1.FavoriteSong
	favoriteSongEntity.Fill(&favoriteSong)

	return &favoriteSong, nil
}

func (s *Service) DeleteFavoriteSong(externalTrn storm.Node, favoriteSongId restApiV1.FavoriteSongId) (*restApiV1.FavoriteSong, error) {
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

	var favoriteSongEntity entity.FavoriteSongEntity
	e = txn.One("Id", string(favoriteSongId.UserId)+":"+string(favoriteSongId.SongId), &favoriteSongEntity)
	if e != nil {
		return nil, e
	}

	// Delete favoriteSong
	e = txn.DeleteStruct(&favoriteSongEntity)
	if e != nil {
		return nil, e
	}

	// Archive favoriteSongId deletion
	e = txn.Save(entity.NewDeletedFavoriteSongEntity(favoriteSongId))
	if e != nil {
		return nil, e
	}

	// Force resync on linked favoritePlaylist
	e = s.updateFavoritePlaylistsContainingSong(txn, favoriteSongId.UserId, favoriteSongId.SongId)
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var favoriteSong restApiV1.FavoriteSong
	favoriteSongEntity.Fill(&favoriteSong)

	return &favoriteSong, nil
}

func (s *Service) GetDeletedFavoriteSongIds(externalTrn storm.Node, fromTs int64) ([]restApiV1.FavoriteSongId, error) {
	if s.ServerConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "GetDeletedFavoriteSongIds")
	}

	var e error

	favoriteSongIds := []restApiV1.FavoriteSongId{}
	deletedFavoriteSongEntities := []entity.DeletedFavoriteSongEntity{}

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(false)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	e = txn.Range("DeleteTs", fromTs, time.Now().UnixNano(), &deletedFavoriteSongEntities)

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	for _, deletedFavoriteSongEntity := range deletedFavoriteSongEntities {
		favoriteSongIds = append(favoriteSongIds, restApiV1.FavoriteSongId{UserId: deletedFavoriteSongEntity.UserId, SongId: deletedFavoriteSongEntity.SongId})
	}

	return favoriteSongIds, nil
}

func (s *Service) GetDeletedUserFavoriteSongIds(externalTrn storm.Node, fromTs int64, userId restApiV1.UserId) ([]restApiV1.SongId, error) {
	var e error

	songIds := []restApiV1.SongId{}
	deletedFavoriteSongEntities := []entity.DeletedFavoriteSongEntity{}

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(false)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	e = txn.Range("DeleteTs", fromTs, time.Now().UnixNano(), &deletedFavoriteSongEntities)

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	for _, deletedFavoriteSongEntity := range deletedFavoriteSongEntities {
		if deletedFavoriteSongEntity.UserId == userId {
			songIds = append(songIds, deletedFavoriteSongEntity.SongId)
		}
	}

	return songIds, nil
}
