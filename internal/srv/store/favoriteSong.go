package store

import (
	"github.com/asdine/storm/v3"
	"github.com/jmoiron/sqlx"
	"github.com/jypelle/mifasol/internal/srv/oldentity"
	"github.com/jypelle/mifasol/internal/srv/storeerror"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"time"
)

func (s *Store) ReadFavoriteSongs(externalTrn *sqlx.Tx, filter *restApiV1.FavoriteSongFilter) ([]restApiV1.FavoriteSong, error) {
	if s.serverConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "ReadFavoriteSongs")
	}

	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.Db.Begin(false)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	favoriteSongEntities := []oldentity.FavoriteSongEntity{}

	if filter.FromTs != nil {
		err = txn.Range("UpdateTs", *filter.FromTs, time.Now().UnixNano(), &favoriteSongEntities)
	} else {
		err = txn.All(&favoriteSongEntities)
	}

	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	favoriteSongs := []restApiV1.FavoriteSong{}

	for _, favoriteSongEntity := range favoriteSongEntities {
		var favoriteSong restApiV1.FavoriteSong
		favoriteSongEntity.Fill(&favoriteSong)
		favoriteSongs = append(favoriteSongs, favoriteSong)
	}

	return favoriteSongs, nil
}

func (s *Store) ReadFavoriteSong(externalTrn *sqlx.Tx, favoriteSongId restApiV1.FavoriteSongId) (*restApiV1.FavoriteSong, error) {
	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.Db.Begin(false)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	var favoriteSongEntity oldentity.FavoriteSongEntity
	err = txn.One("Id", string(favoriteSongId.UserId)+":"+string(favoriteSongId.SongId), &favoriteSongEntity)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, storeerror.ErrNotFound
		}
		return nil, err
	}

	var favoriteSong restApiV1.FavoriteSong
	favoriteSongEntity.Fill(&favoriteSong)

	return &favoriteSong, nil
}

func (s *Store) CreateFavoriteSong(externalTrn *sqlx.Tx, favoriteSongMeta *restApiV1.FavoriteSongMeta, check bool) (*restApiV1.FavoriteSong, error) {
	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	var favoriteSongEntity oldentity.FavoriteSongEntity

	err = txn.One("Id", string(favoriteSongMeta.Id.UserId)+":"+string(favoriteSongMeta.Id.SongId), &favoriteSongEntity)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}
	if err == storm.ErrNotFound {
		// Store favorite song
		now := time.Now().UnixNano()

		favoriteSongEntity = oldentity.FavoriteSongEntity{
			UpdateTs: now,
		}
		favoriteSongEntity.LoadMeta(favoriteSongMeta)

		err = txn.Save(&favoriteSongEntity)
		if err != nil {
			return nil, err
		}

		// if previously deletedFavoriteSong exists
		var deletedFavoriteSongEntity oldentity.DeletedFavoriteSongEntity
		err = txn.One("Id", string(favoriteSongMeta.Id.UserId)+":"+string(favoriteSongMeta.Id.SongId), &deletedFavoriteSongEntity)
		if err != nil && err != storm.ErrNotFound {
			return nil, err
		}

		if err == nil {
			// Delete deletedFavoriteSong
			err = txn.DeleteStruct(&deletedFavoriteSongEntity)
			if err != nil {
				return nil, err
			}
		}

		// Force resync on linked favoritePlaylist
		err = s.updateFavoritePlaylistsContainingSong(txn, favoriteSongMeta.Id.UserId, favoriteSongMeta.Id.SongId)
		if err != nil {
			return nil, err
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

func (s *Store) DeleteFavoriteSong(externalTrn *sqlx.Tx, favoriteSongId restApiV1.FavoriteSongId) (*restApiV1.FavoriteSong, error) {
	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	var favoriteSongEntity oldentity.FavoriteSongEntity
	err = txn.One("Id", string(favoriteSongId.UserId)+":"+string(favoriteSongId.SongId), &favoriteSongEntity)
	if err != nil {
		return nil, err
	}

	// Delete favoriteSong
	err = txn.DeleteStruct(&favoriteSongEntity)
	if err != nil {
		return nil, err
	}

	// Archive favoriteSongId deletion
	err = txn.Save(oldentity.NewDeletedFavoriteSongEntity(favoriteSongId))
	if err != nil {
		return nil, err
	}

	// Force resync on linked favoritePlaylist
	err = s.updateFavoritePlaylistsContainingSong(txn, favoriteSongId.UserId, favoriteSongId.SongId)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var favoriteSong restApiV1.FavoriteSong
	favoriteSongEntity.Fill(&favoriteSong)

	return &favoriteSong, nil
}

func (s *Store) GetDeletedFavoriteSongIds(externalTrn *sqlx.Tx, fromTs int64) ([]restApiV1.FavoriteSongId, error) {
	if s.serverConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "GetDeletedFavoriteSongIds")
	}

	var err error

	favoriteSongIds := []restApiV1.FavoriteSongId{}
	deletedFavoriteSongEntities := []oldentity.DeletedFavoriteSongEntity{}

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	err = txn.Range("DeleteTs", fromTs, time.Now().UnixNano(), &deletedFavoriteSongEntities)

	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	for _, deletedFavoriteSongEntity := range deletedFavoriteSongEntities {
		favoriteSongIds = append(favoriteSongIds, restApiV1.FavoriteSongId{UserId: deletedFavoriteSongEntity.UserId, SongId: deletedFavoriteSongEntity.SongId})
	}

	return favoriteSongIds, nil
}

func (s *Store) GetDeletedUserFavoriteSongIds(externalTrn *sqlx.Tx, fromTs int64, userId restApiV1.UserId) ([]restApiV1.SongId, error) {
	var err error

	songIds := []restApiV1.SongId{}
	deletedFavoriteSongEntities := []oldentity.DeletedFavoriteSongEntity{}

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	err = txn.Range("DeleteTs", fromTs, time.Now().UnixNano(), &deletedFavoriteSongEntities)

	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	for _, deletedFavoriteSongEntity := range deletedFavoriteSongEntities {
		if deletedFavoriteSongEntity.UserId == userId {
			songIds = append(songIds, deletedFavoriteSongEntity.SongId)
		}
	}

	return songIds, nil
}
