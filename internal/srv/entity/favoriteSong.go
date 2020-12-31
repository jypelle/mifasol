package entity

import (
	"github.com/jypelle/mifasol/restApiV1"
	"time"
)

type FavoriteSongEntity struct {
	Id       string           `storm:"id"`
	UpdateTs int64            `storm:"index"`
	UserId   restApiV1.UserId `storm:"index"`
	SongId   restApiV1.SongId `storm:"index"`
}

func (e *FavoriteSongEntity) Fill(s *restApiV1.FavoriteSong) {
	s.Id = restApiV1.FavoriteSongId{UserId: e.UserId, SongId: e.SongId}
	s.UpdateTs = e.UpdateTs
}

func (e *FavoriteSongEntity) LoadMeta(s *restApiV1.FavoriteSongMeta) {
	if s != nil {
		e.Id = string(s.Id.UserId) + ":" + string(s.Id.SongId)
		e.UserId = s.Id.UserId
		e.SongId = s.Id.SongId
	}
}

type DeletedFavoriteSongEntity struct {
	Id       string           `storm:"id"`
	DeleteTs int64            `storm:"index"`
	UserId   restApiV1.UserId `storm:"index"`
	SongId   restApiV1.SongId `storm:"index"`
}

func NewDeletedFavoriteSongEntity(favoriteSongId restApiV1.FavoriteSongId) *DeletedFavoriteSongEntity {
	now := time.Now().UnixNano()

	return &DeletedFavoriteSongEntity{
		Id:       string(favoriteSongId.UserId) + ":" + string(favoriteSongId.SongId),
		DeleteTs: now,
		UserId:   favoriteSongId.UserId,
		SongId:   favoriteSongId.SongId,
	}
}
