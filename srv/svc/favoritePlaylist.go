package svc

import (
	"github.com/asdine/storm/v3"
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

	favoritePlaylistEntities := []entity.FavoritePlaylistEntity{}

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
	e = txn.One("Id", string(favoritePlaylistId.UserId)+":"+string(favoritePlaylistId.PlaylistId), &favoritePlaylistEntity)
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

	e = txn.One("Id", string(favoritePlaylistMeta.Id.UserId)+":"+string(favoritePlaylistMeta.Id.PlaylistId), &favoritePlaylistEntity)
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
		e = txn.One("Id", string(favoritePlaylistMeta.Id.UserId)+":"+string(favoritePlaylistMeta.Id.PlaylistId), &deletedFavoritePlaylistEntity)
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

		// Add favorite playlist songs to favorite songs
		playlistSongEntities := []entity.PlaylistSongEntity{}

		e = txn.Find("PlaylistId", favoritePlaylistMeta.Id.PlaylistId, &playlistSongEntities)
		if e != nil && e != storm.ErrNotFound {
			return nil, e
		}

		for _, playlistSongEntity := range playlistSongEntities {
			_, e = s.CreateFavoriteSong(txn, &restApiV1.FavoriteSongMeta{Id: restApiV1.FavoriteSongId{UserId: favoritePlaylistMeta.Id.UserId, SongId: playlistSongEntity.SongId}}, false)
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
	e = txn.One("Id", string(favoritePlaylistId.UserId)+":"+string(favoritePlaylistId.PlaylistId), &favoritePlaylistEntity)
	if e != nil {
		return nil, e
	}

	// Delete favoritePlaylist
	e = txn.DeleteStruct(&favoritePlaylistEntity)
	if e != nil {
		return nil, e
	}

	// Archive favoritePlaylistId deletion
	e = txn.Save(entity.NewDeletedFavoritePlaylistEntity(favoritePlaylistId))
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

	e = txn.Range("DeleteTs", fromTs, time.Now().UnixNano(), &deletedFavoritePlaylistEntities)

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	for _, deletedFavoritePlaylistEntity := range deletedFavoritePlaylistEntities {
		favoritePlaylistIds = append(favoritePlaylistIds, restApiV1.FavoritePlaylistId{UserId: deletedFavoritePlaylistEntity.UserId, PlaylistId: deletedFavoritePlaylistEntity.PlaylistId})
	}

	return favoritePlaylistIds, nil
}

func (s *Service) GetDeletedUserFavoritePlaylistIds(externalTrn storm.Node, fromTs int64, userId restApiV1.UserId) ([]restApiV1.PlaylistId, error) {
	var e error

	playlistIds := []restApiV1.PlaylistId{}
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

	e = txn.Range("DeleteTs", fromTs, time.Now().UnixNano(), &deletedFavoritePlaylistEntities)

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	for _, deletedFavoritePlaylistEntity := range deletedFavoritePlaylistEntities {
		if deletedFavoritePlaylistEntity.UserId == userId {
			playlistIds = append(playlistIds, deletedFavoritePlaylistEntity.PlaylistId)
		}
	}

	return playlistIds, nil
}

func (s *Service) updateFavoritePlaylistsContainingSong(txn storm.Node, userId restApiV1.UserId, songId restApiV1.SongId) error {
	now := time.Now().UnixNano()

	favoritePlaylistEntities := []entity.FavoritePlaylistEntity{}

	e := txn.Find("UserId", userId, &favoritePlaylistEntities)
	if e != nil && e != storm.ErrNotFound {
		return e
	}

	for _, favoritePlaylistEntity := range favoritePlaylistEntities {
		var playlistSongEntity entity.PlaylistSongEntity
		e = txn.One("Id", string(favoritePlaylistEntity.PlaylistId)+":"+string(songId), &playlistSongEntity)
		if e != nil && e != storm.ErrNotFound {
			return e
		}
		if e != storm.ErrNotFound {
			favoritePlaylistEntity.UpdateTs = now

			e = txn.Save(&favoritePlaylistEntity)
			if e != nil {
				return e
			}
		}
	}
	return nil
}
