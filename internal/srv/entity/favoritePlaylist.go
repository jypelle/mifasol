package entity

import (
	"github.com/jypelle/mifasol/restApiV1"
	"time"
)

type FavoritePlaylistEntity struct {
	UserId     restApiV1.UserId     `db:"user_id"`
	PlaylistId restApiV1.PlaylistId `db:"playlist_id"`
	UpdateTs   int64                `db:"update_ts"`
}

func (e *FavoritePlaylistEntity) Fill(s *restApiV1.FavoritePlaylist) {
	s.Id = restApiV1.FavoritePlaylistId{UserId: e.UserId, PlaylistId: e.PlaylistId}
	s.UpdateTs = e.UpdateTs
}

func (e *FavoritePlaylistEntity) LoadMeta(s *restApiV1.FavoritePlaylistMeta) {
	if s != nil {
		e.UserId = s.Id.UserId
		e.PlaylistId = s.Id.PlaylistId
	}
}

type DeletedFavoritePlaylistEntity struct {
	UserId     restApiV1.UserId     `db:"user_id"`
	PlaylistId restApiV1.PlaylistId `db:"playlist_id"`
	DeleteTs   int64                `db:"delete_ts"`
}

func NewDeletedFavoritePlaylistEntity(favoritePlaylistId restApiV1.FavoritePlaylistId) *DeletedFavoritePlaylistEntity {
	now := time.Now().UnixNano()

	return &DeletedFavoritePlaylistEntity{
		DeleteTs:   now,
		UserId:     favoritePlaylistId.UserId,
		PlaylistId: favoritePlaylistId.PlaylistId,
	}
}
