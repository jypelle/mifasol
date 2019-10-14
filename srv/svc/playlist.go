package svc

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"lyra/restApiV1"
	"lyra/srv/entity"
	"lyra/tool"
	"reflect"
	"sort"
	"time"
)

func (s *Service) ReadPlaylists(externalTrn storm.Node, filter *restApiV1.PlaylistFilter) ([]restApiV1.Playlist, error) {
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
	if filter.ContentFromTs != nil {
		matchers = append(matchers, q.Gte("ContentUpdateTs", *filter.ContentFromTs))
	}

	query := txn.Select(matchers...)

	switch filter.Order {
	case restApiV1.PlaylistOrderByPlaylistName:
		query = query.OrderBy("Name")
	case restApiV1.PlaylistOrderByUpdateTs:
		query = query.OrderBy("UpdateTs")
	case restApiV1.PlaylistOrderByContentUpdateTs:
		query = query.OrderBy("ContentUpdateTs")
	default:
	}

	playlistEntities := []entity.PlaylistEntity{}

	e = query.Find(&playlistEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	playlists := []restApiV1.Playlist{}

	for _, playlistEntity := range playlistEntities {
		var playlist restApiV1.Playlist
		playlistEntity.Fill(&playlist)
		playlists = append(playlists, playlist)
	}

	return playlists, nil
}

func (s *Service) ReadPlaylist(externalTrn storm.Node, playlistId string) (*restApiV1.Playlist, error) {
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

	var playlistEntity entity.PlaylistEntity
	e = txn.One("Id", playlistId, &playlistEntity)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}

	var playlist restApiV1.Playlist
	playlistEntity.Fill(&playlist)

	return &playlist, nil
}

func (s *Service) CreatePlaylist(externalTrn storm.Node, playlistMeta *restApiV1.PlaylistMeta, check bool) (*restApiV1.Playlist, error) {
	return s.CreateInternalPlaylist(externalTrn, "", playlistMeta, check)
}

func (s *Service) CreateInternalPlaylist(externalTrn storm.Node, playlistId string, playlistMeta *restApiV1.PlaylistMeta, check bool) (*restApiV1.Playlist, error) {
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
		playlistId = tool.CreateUlid()
	}

	playlistEntity := entity.PlaylistEntity{
		Id:              playlistId,
		CreationTs:      now,
		UpdateTs:        now,
		ContentUpdateTs: now,
	}
	playlistEntity.LoadMeta(playlistMeta)

	// Clean owner list
	playlistEntity.OwnerUserIds = tool.Deduplicate(playlistEntity.OwnerUserIds)
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
			var userEntity entity.UserEntity
			e = txn.One("Id", ownerUserId, &userEntity)
			if e != nil {
				return nil, e
			}
		}

		// Store playlist owner
		e = txn.Save(entity.NewOwnedUserPlaylistEntity(ownerUserId, playlistId))
		if e != nil {
			return nil, e
		}
	}

	// Create songs link
	for _, songId := range playlistEntity.SongIds {
		// Check song id
		if check {
			var songEntity entity.SongEntity
			e = txn.One("Id", songId, &songEntity)
			if e != nil {
				return nil, e
			}
		}

		// Store song link
		e = txn.Save(entity.NewPlaylistSongEntity(playlistId, songId))
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

func (s *Service) UpdatePlaylist(externalTrn storm.Node, playlistId string, playlistMeta *restApiV1.PlaylistMeta, check bool) (*restApiV1.Playlist, error) {
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

	var playlistEntity entity.PlaylistEntity
	e = txn.One("Id", playlistId, &playlistEntity)
	if e != nil {
		return nil, e
	}

	playlistOldName := playlistEntity.Name
	playlistOldSongIds := playlistEntity.SongIds
	playlistOldOwnerUserIds := playlistEntity.OwnerUserIds

	playlistEntity.LoadMeta(playlistMeta)

	// Clean owner list
	playlistEntity.OwnerUserIds = tool.Deduplicate(playlistEntity.OwnerUserIds)
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

	e = txn.Update(&playlistEntity)
	if e != nil {
		return nil, e
	}

	// Update owner index
	for _, ownerUserId := range playlistOldOwnerUserIds {
		query := txn.Select(q.Eq("UserId", ownerUserId))
		e = query.Delete(new(entity.OwnedUserPlaylistEntity))
		if e != nil && e != storm.ErrNotFound {
			return nil, e
		}
	}

	for _, ownerUserId := range playlistEntity.OwnerUserIds {

		// Check owner user id
		if check {
			var userEntity entity.UserEntity
			e := txn.One("Id", ownerUserId, &userEntity)
			if e != nil {
				return nil, e
			}
		}

		// Store playlist owner
		e = txn.Save(entity.NewOwnedUserPlaylistEntity(ownerUserId, playlistId))
		if e != nil {
			return nil, e
		}
	}

	// Update songs list
	if songIdsUpdated {
		for _, songId := range playlistOldSongIds {
			query := txn.Select(q.Eq("SongId", songId))
			e = query.Delete(new(entity.PlaylistSongEntity))
			if e != nil && e != storm.ErrNotFound {
				return nil, e
			}
		}
		for _, songId := range playlistEntity.SongIds {
			// Check song id
			if check {
				var song restApiV1.Song
				e := txn.One("Id", songId, &song)
				if e != nil {
					return nil, e
				}
			}

			// Store song link
			e = txn.Save(entity.NewPlaylistSongEntity(playlistId, songId))
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

func (s *Service) AddSongToPlaylist(externalTrn storm.Node, playlistId string, songId string, check bool) (*restApiV1.Playlist, error) {
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

	var playlistEntity entity.PlaylistEntity
	e = txn.One("Id", playlistId, &playlistEntity)
	if e != nil {
		return nil, e
	}

	// Check song id
	if check {
		var song restApiV1.Song
		e = txn.One("Id", songId, &song)
		if e != nil {
			return nil, e
		}
	}

	playlistEntity.SongIds = append(playlistEntity.SongIds, songId)

	// Update playlist update timestamp
	playlistEntity.UpdateTs = now
	playlistEntity.ContentUpdateTs = now

	e = txn.Update(&playlistEntity)
	if e != nil {
		return nil, e
	}

	// Store song link
	e = txn.Save(entity.NewPlaylistSongEntity(playlistId, songId))
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

func (s *Service) DeletePlaylist(externalTrn storm.Node, playlistId string) (*restApiV1.Playlist, error) {
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

	var playlistEntity entity.PlaylistEntity
	e = txn.One("Id", playlistId, &playlistEntity)
	if e != nil {
		return nil, e
	}

	// Delete ower link
	for _, ownerUserId := range playlistEntity.OwnerUserIds {
		query := txn.Select(q.Eq("UserId", ownerUserId))
		e = query.Delete(new(entity.OwnedUserPlaylistEntity))
		if e != nil && e != storm.ErrNotFound {
			return nil, e
		}
	}

	// Delete songs link
	for _, _songId := range playlistEntity.SongIds {
		query := txn.Select(q.Eq("SongId", _songId))
		e = query.Delete(new(entity.PlaylistSongEntity))
		if e != nil && e != storm.ErrNotFound {
			return nil, e
		}
	}

	// Delete playlist
	e = txn.DeleteStruct(&playlistEntity)
	if e != nil {
		return nil, e
	}

	// Archive playlistId
	e = txn.Save(&entity.DeletedPlaylistEntity{Id: playlistEntity.Id, DeleteTs: deleteTs})
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

func (s *Service) GetDeletedPlaylistIds(externalTrn storm.Node, fromTs int64) ([]string, error) {
	var e error

	playlistIds := []string{}
	deletedPlaylistEntities := []entity.DeletedPlaylistEntity{}

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

	e = query.Find(&deletedPlaylistEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	for _, deletedPlaylistEntity := range deletedPlaylistEntities {
		playlistIds = append(playlistIds, deletedPlaylistEntity.Id)
	}

	return playlistIds, nil
}

func (s *Service) GetPlaylistIdsFromSongId(externalTrn storm.Node, songId string) ([]string, error) {
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

	var playlistIds []string
	playlistSongEntities := []entity.PlaylistSongEntity{}

	query := txn.Select(q.Eq("SongId", songId))

	e = query.Find(&playlistSongEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	for _, playlistSongEntity := range playlistSongEntities {
		playlistIds = append(playlistIds, playlistSongEntity.Id.PlaylistId)
	}

	return playlistIds, nil

}

func (s *Service) GetPlaylistIdsFromOwnerUserId(externalTrn storm.Node, ownerUserId string) ([]string, error) {
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

	var playlistIds []string
	ownedUserPlaylistEntities := []entity.OwnedUserPlaylistEntity{}

	query := txn.Select(q.Eq("UserId", ownerUserId))
	e = query.Find(&ownedUserPlaylistEntities)

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	for _, ownedUserPlaylistEntity := range ownedUserPlaylistEntities {
		playlistIds = append(playlistIds, ownedUserPlaylistEntity.Id.PlaylistId)
	}

	return playlistIds, nil

}
