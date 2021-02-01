package oldstore

import (
	"errors"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/jypelle/mifasol/internal/srv/legacy/oldentity"
	"github.com/jypelle/mifasol/internal/srv/storeerror"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"reflect"
	"sort"
	"time"
)

func (s *OldStore) ReadPlaylists(externalTrn storm.Node, filter *restApiV1.PlaylistFilter) ([]restApiV1.Playlist, error) {
	if s.ServerConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "ReadPlaylists")
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

	playlistEntities := []oldentity.PlaylistEntity{}

	if filter.FromTs != nil {
		e = txn.Range("UpdateTs", *filter.FromTs, time.Now().UnixNano(), &playlistEntities)
	} else {
		e = txn.All(&playlistEntities)
	}

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	playlists := []restApiV1.Playlist{}

	for _, playlistEntity := range playlistEntities {
		if filter.FavoriteUserId != nil {
			fav, e := s.ReadFavoritePlaylist(txn, restApiV1.FavoritePlaylistId{UserId: *filter.FavoriteUserId, PlaylistId: playlistEntity.Id})
			if e != nil {
				continue
			}
			if filter.FavoriteFromTs != nil {
				if playlistEntity.ContentUpdateTs < *filter.FavoriteFromTs && fav.UpdateTs < *filter.FavoriteFromTs {
					continue
				}
			}
		}

		var playlist restApiV1.Playlist
		playlistEntity.Fill(&playlist)
		playlists = append(playlists, playlist)
	}

	return playlists, nil
}

func (s *OldStore) ReadPlaylist(externalTrn storm.Node, playlistId restApiV1.PlaylistId) (*restApiV1.Playlist, error) {
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

	var playlistEntity oldentity.PlaylistEntity
	e = txn.One("Id", playlistId, &playlistEntity)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, storeerror.ErrNotFound
		}
		return nil, e
	}

	var playlist restApiV1.Playlist
	playlistEntity.Fill(&playlist)

	return &playlist, nil
}

func (s *OldStore) CreatePlaylist(externalTrn storm.Node, playlistMeta *restApiV1.PlaylistMeta, check bool) (*restApiV1.Playlist, error) {
	return s.CreateInternalPlaylist(externalTrn, "", playlistMeta, check)
}

func (s *OldStore) CreateInternalPlaylist(externalTrn storm.Node, playlistId restApiV1.PlaylistId, playlistMeta *restApiV1.PlaylistMeta, check bool) (*restApiV1.Playlist, error) {
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

	if playlistId == "" {
		playlistId = restApiV1.PlaylistId(tool.CreateUlid())
	}

	playlistEntity := oldentity.PlaylistEntity{
		Id:              playlistId,
		CreationTs:      now,
		UpdateTs:        now,
		ContentUpdateTs: now,
	}
	playlistEntity.LoadMeta(playlistMeta)

	// Clean owner list
	playlistEntity.OwnerUserIds = tool.DeduplicateUserId(playlistEntity.OwnerUserIds)
	sort.Slice(playlistEntity.OwnerUserIds, func(i, j int) bool {
		return playlistEntity.OwnerUserIds[i] < playlistEntity.OwnerUserIds[j]
	})

	e = txn.Save(&playlistEntity)
	if e != nil {
		return nil, e
	}

	// Create owners link
	for _, ownerUserId := range playlistEntity.OwnerUserIds {
		// Check owner user id
		if check {
			var userEntity oldentity.UserEntity
			e = txn.One("Id", ownerUserId, &userEntity)
			if e != nil {
				return nil, e
			}
		}

		// Store playlist owner
		e = txn.Save(oldentity.NewOwnedUserPlaylistEntity(ownerUserId, playlistId))
		if e != nil {
			return nil, e
		}

		// Add playlist to owner favorite playlist
		favoritePlaylistMeta := &restApiV1.FavoritePlaylistMeta{restApiV1.FavoritePlaylistId{UserId: ownerUserId, PlaylistId: playlistId}}
		_, e = s.CreateFavoritePlaylist(txn, favoritePlaylistMeta, false)
		if e != nil {
			return nil, e
		}
	}

	// Create songs link
	for _, songId := range playlistEntity.SongIds {
		// Check song id
		if check {
			var songEntity oldentity.SongEntity
			e = txn.One("Id", songId, &songEntity)
			if e != nil {
				return nil, e
			}
		}

		// Store song link
		e = txn.Save(oldentity.NewPlaylistSongEntity(playlistId, songId))
		if e != nil {
			return nil, e
		}
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var playlist restApiV1.Playlist
	playlistEntity.Fill(&playlist)

	return &playlist, nil
}

func (s *OldStore) UpdatePlaylist(externalTrn storm.Node, playlistId restApiV1.PlaylistId, playlistMeta *restApiV1.PlaylistMeta, check bool) (*restApiV1.Playlist, error) {
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

	now := time.Now().UnixNano()

	var playlistEntity oldentity.PlaylistEntity
	e = txn.One("Id", playlistId, &playlistEntity)
	if e != nil {
		return nil, e
	}

	playlistOldName := playlistEntity.Name
	playlistOldSongIds := playlistEntity.SongIds

	playlistEntity.LoadMeta(playlistMeta)

	// Clean owner list
	playlistEntity.OwnerUserIds = tool.DeduplicateUserId(playlistEntity.OwnerUserIds)
	sort.Slice(playlistEntity.OwnerUserIds, func(i, j int) bool {
		return playlistEntity.OwnerUserIds[i] < playlistEntity.OwnerUserIds[j]
	})

	// Detect song list update
	songIdsUpdated := !reflect.DeepEqual(playlistOldSongIds, playlistEntity.SongIds)

	// Update playlist update timestamp
	playlistEntity.UpdateTs = now

	// Update playlist update content timestamp
	if playlistOldName != playlistEntity.Name || songIdsUpdated {
		playlistEntity.ContentUpdateTs = now
	}

	e = txn.Save(&playlistEntity)
	if e != nil {
		return nil, e
	}

	// Update owner index
	e = txn.Select(q.Eq("PlaylistId", playlistId)).Delete(&oldentity.OwnedUserPlaylistEntity{})
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	for _, ownerUserId := range playlistEntity.OwnerUserIds {

		// Check owner user id
		if check {
			var userEntity oldentity.UserEntity
			e := txn.One("Id", ownerUserId, &userEntity)
			if e != nil {
				return nil, e
			}
		}

		// Store playlist owner
		e = txn.Save(oldentity.NewOwnedUserPlaylistEntity(ownerUserId, playlistId))
		if e != nil {
			return nil, e
		}
	}

	// Update songs list
	if songIdsUpdated {
		e = txn.Select(q.Eq("PlaylistId", playlistId)).Delete(&oldentity.PlaylistSongEntity{})
		if e != nil && e != storm.ErrNotFound {
			return nil, e
		}

		for _, songId := range playlistEntity.SongIds {
			// Check song id
			if check {
				var songEntity oldentity.SongEntity
				e := txn.One("Id", songId, &songEntity)
				if e != nil {
					return nil, e
				}
			}

			// Store song link
			e = txn.Save(oldentity.NewPlaylistSongEntity(playlistId, songId))
			if e != nil {
				return nil, e
			}
		}
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var playlist restApiV1.Playlist
	playlistEntity.Fill(&playlist)

	return &playlist, nil
}

func (s *OldStore) AddSongToPlaylist(externalTrn storm.Node, playlistId restApiV1.PlaylistId, songId restApiV1.SongId, check bool) (*restApiV1.Playlist, error) {
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

	now := time.Now().UnixNano()

	var playlistEntity oldentity.PlaylistEntity
	e = txn.One("Id", playlistId, &playlistEntity)
	if e != nil {
		return nil, e
	}

	// Check song id
	if check {
		var songEntity oldentity.SongEntity
		e = txn.One("Id", songId, &songEntity)
		if e != nil {
			return nil, e
		}
	}

	playlistEntity.SongIds = append(playlistEntity.SongIds, songId)

	if playlistId == restApiV1.IncomingPlaylistId && len(playlistEntity.SongIds) > 100 {
		var songIdToRemove restApiV1.SongId
		songIdToRemove, playlistEntity.SongIds = playlistEntity.SongIds[0], playlistEntity.SongIds[1:]
		if !tool.ContainsSongId(playlistEntity.SongIds, songIdToRemove) {
			e = txn.DeleteStruct(oldentity.NewPlaylistSongEntity(playlistId, songIdToRemove))
			if e != nil {
				return nil, e
			}
		}
	}

	// Update playlist update timestamp
	playlistEntity.UpdateTs = now
	playlistEntity.ContentUpdateTs = now

	e = txn.Save(&playlistEntity)
	if e != nil {
		return nil, e
	}

	// Store song link
	e = txn.Save(oldentity.NewPlaylistSongEntity(playlistId, songId))
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var playlist restApiV1.Playlist
	playlistEntity.Fill(&playlist)

	return &playlist, nil
}

func (s *OldStore) DeletePlaylist(externalTrn storm.Node, playlistId restApiV1.PlaylistId) (*restApiV1.Playlist, error) {
	var e error

	// Incoming playlist can't be deleted
	if playlistId == restApiV1.IncomingPlaylistId {
		return nil, errors.New("Incoming playlist can't be deleted")
	}

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

	var playlistEntity oldentity.PlaylistEntity
	e = txn.One("Id", playlistId, &playlistEntity)
	if e != nil {
		return nil, e
	}

	// Delete favorite playlist link
	favoritePlaylistEntities := []oldentity.FavoritePlaylistEntity{}
	e = txn.Find("PlaylistId", playlistId, &favoritePlaylistEntities)

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}
	for _, favoritePlaylistEntity := range favoritePlaylistEntities {
		s.DeleteFavoritePlaylist(txn, restApiV1.FavoritePlaylistId{UserId: favoritePlaylistEntity.UserId, PlaylistId: favoritePlaylistEntity.PlaylistId})
	}

	// Delete owners link
	ownedUserPlaylistEntities := []oldentity.OwnedUserPlaylistEntity{}
	e = txn.Find("PlaylistId", playlistId, &ownedUserPlaylistEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}
	for _, ownedUserPlaylistEntity := range ownedUserPlaylistEntities {
		txn.DeleteStruct(&ownedUserPlaylistEntity)
	}

	// Delete songs link
	playlistSongEntities := []oldentity.PlaylistSongEntity{}
	e = txn.Find("PlaylistId", playlistId, &playlistSongEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}
	for _, playlistSongEntity := range playlistSongEntities {
		txn.DeleteStruct(&playlistSongEntity)
	}

	// Delete playlist
	e = txn.DeleteStruct(&playlistEntity)
	if e != nil {
		return nil, e
	}

	// Archive playlistId
	e = txn.Save(&oldentity.DeletedPlaylistEntity{Id: playlistEntity.Id, DeleteTs: deleteTs})
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var playlist restApiV1.Playlist
	playlistEntity.Fill(&playlist)

	return &playlist, nil
}

func (s *OldStore) GetDeletedPlaylistIds(externalTrn storm.Node, fromTs int64) ([]restApiV1.PlaylistId, error) {
	if s.ServerConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "GetDeletedPlaylistIds")
	}

	var e error

	playlistIds := []restApiV1.PlaylistId{}
	deletedPlaylistEntities := []oldentity.DeletedPlaylistEntity{}

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(false)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	e = txn.Range("DeleteTs", fromTs, time.Now().UnixNano(), &deletedPlaylistEntities)

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	for _, deletedPlaylistEntity := range deletedPlaylistEntities {
		playlistIds = append(playlistIds, deletedPlaylistEntity.Id)
	}

	return playlistIds, nil
}

func (s *OldStore) GetPlaylistIdsFromSongId(externalTrn storm.Node, songId restApiV1.SongId) ([]restApiV1.PlaylistId, error) {
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

	var playlistIds []restApiV1.PlaylistId
	playlistSongEntities := []oldentity.PlaylistSongEntity{}

	e = txn.Find("SongId", songId, &playlistSongEntities)

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	for _, playlistSongEntity := range playlistSongEntities {
		playlistIds = append(playlistIds, playlistSongEntity.PlaylistId)
	}

	return playlistIds, nil

}

func (s *OldStore) GetPlaylistIdsFromOwnerUserId(externalTrn storm.Node, ownerUserId restApiV1.UserId) ([]restApiV1.PlaylistId, error) {
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

	var playlistIds []restApiV1.PlaylistId
	ownedUserPlaylistEntities := []oldentity.OwnedUserPlaylistEntity{}

	e = txn.Find("UserId", ownerUserId, &ownedUserPlaylistEntities)

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	for _, ownedUserPlaylistEntity := range ownedUserPlaylistEntities {
		playlistIds = append(playlistIds, ownedUserPlaylistEntity.PlaylistId)
	}

	return playlistIds, nil

}
