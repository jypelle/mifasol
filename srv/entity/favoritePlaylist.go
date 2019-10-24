package entity

import (
	"github.com/jypelle/mifasol/restApiV1"
	"time"
)

type FavoritePlaylistEntity struct {
	Id         string `storm:"id"`
	UpdateTs   int64  `storm:"index"`
	UserId     string `storm:"index"`
	PlaylistId string `storm:"index"`
}

func NewFavoritePlaylistEntity(userId string, playlistId string) *FavoritePlaylistEntity {
	now := time.Now().UnixNano()

	return &FavoritePlaylistEntity{
		Id:         userId + ":" + playlistId,
		UpdateTs:   now,
		UserId:     userId,
		PlaylistId: playlistId,
	}
}

func (e *FavoritePlaylistEntity) Fill(s *restApiV1.FavoritePlaylist) {
	s.Id = restApiV1.FavoritePlaylistId{UserId: e.UserId, PlaylistId: e.PlaylistId}
	s.UpdateTs = e.UpdateTs
}

func (e *FavoritePlaylistEntity) LoadMeta(s *restApiV1.FavoritePlaylistMeta) {
	if s != nil {
		e.Id = s.Id.UserId + ":" + s.Id.PlaylistId
		e.UserId = s.Id.UserId
		e.PlaylistId = s.Id.PlaylistId
	}
}

type DeletedFavoritePlaylistEntity struct {
	Id         string `storm:"id"`
	DeleteTs   int64  `storm:"index"`
	UserId     string `storm:"index"`
	PlaylistId string `storm:"index"`
}

func NewDeletedFavoritePlaylistEntity(favoritePlaylistId restApiV1.FavoritePlaylistId) *DeletedFavoritePlaylistEntity {
	now := time.Now().UnixNano()

	return &DeletedFavoritePlaylistEntity{
		Id:         favoritePlaylistId.UserId + ":" + favoritePlaylistId.PlaylistId,
		DeleteTs:   now,
		UserId:     favoritePlaylistId.UserId,
		PlaylistId: favoritePlaylistId.PlaylistId,
	}
}
