package oldentity

import (
	"github.com/jypelle/mifasol/restApiV1"
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

type DeletedFavoriteSongEntity struct {
	Id       string           `storm:"id"`
	DeleteTs int64            `storm:"index"`
	UserId   restApiV1.UserId `storm:"index"`
	SongId   restApiV1.SongId `storm:"index"`
}
