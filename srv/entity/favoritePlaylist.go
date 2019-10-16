package entity

import (
	"mifasol/restApiV1"
	"time"
)

type FavoritePlaylistEntity struct {
	Id         restApiV1.FavoritePlaylistId `storm:"id"`
	UpdateTs   int64                        `storm:"index"`
	UserId     string                       `storm:"index"`
	PlaylistId string                       `storm:"index"`
}

func NewFavoritePlaylistEntity(userId string, playlistId string) *FavoritePlaylistEntity {
	now := time.Now().UnixNano()

	return &FavoritePlaylistEntity{
		Id: restApiV1.FavoritePlaylistId{
			UserId:     userId,
			PlaylistId: playlistId,
		},
		UpdateTs:   now,
		UserId:     userId,
		PlaylistId: playlistId,
	}
}

func (e *FavoritePlaylistEntity) Fill(s *restApiV1.FavoritePlaylist) {
	s.Id = e.Id
	s.UpdateTs = e.UpdateTs
}

func (e *FavoritePlaylistEntity) LoadMeta(s *restApiV1.FavoritePlaylistMeta) {
	if s != nil {
		e.Id = s.Id
	}
}

type DeletedFavoritePlaylistEntity struct {
	Id       restApiV1.FavoritePlaylistId `storm:"id"`
	DeleteTs int64                        `storm:"index"`
}
