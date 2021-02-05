package entity

import (
	"github.com/jypelle/mifasol/restApiV1"
	"time"
)

type FavoriteSongEntity struct {
	UserId   restApiV1.UserId `db:"user_id"`
	SongId   restApiV1.SongId `db:"song_id"`
	UpdateTs int64            `db:"update_ts"`
}

func (e *FavoriteSongEntity) Fill(s *restApiV1.FavoriteSong) {
	s.Id = restApiV1.FavoriteSongId{UserId: e.UserId, SongId: e.SongId}
	s.UpdateTs = e.UpdateTs
}

func (e *FavoriteSongEntity) LoadMeta(s *restApiV1.FavoriteSongMeta) {
	if s != nil {
		e.UserId = s.Id.UserId
		e.SongId = s.Id.SongId
	}
}

type DeletedFavoriteSongEntity struct {
	UserId   restApiV1.UserId `db:"user_id"`
	SongId   restApiV1.SongId `db:"song_id"`
	DeleteTs int64            `db:"delete_ts"`
}

func NewDeletedFavoriteSongEntity(favoriteSongId restApiV1.FavoriteSongId) *DeletedFavoriteSongEntity {
	now := time.Now().UnixNano()

	return &DeletedFavoriteSongEntity{
		UserId:   favoriteSongId.UserId,
		SongId:   favoriteSongId.SongId,
		DeleteTs: now,
	}
}
