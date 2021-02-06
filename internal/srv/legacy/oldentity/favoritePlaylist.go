package oldentity

import (
	"github.com/jypelle/mifasol/restApiV1"
)

type FavoritePlaylistEntity struct {
	Id         string               `storm:"id"`
	UpdateTs   int64                `storm:"index"`
	UserId     restApiV1.UserId     `storm:"index"`
	PlaylistId restApiV1.PlaylistId `storm:"index"`
}

func (e *FavoritePlaylistEntity) Fill(s *restApiV1.FavoritePlaylist) {
	s.Id = restApiV1.FavoritePlaylistId{UserId: e.UserId, PlaylistId: e.PlaylistId}
	s.UpdateTs = e.UpdateTs
}

type DeletedFavoritePlaylistEntity struct {
	Id         string               `storm:"id"`
	DeleteTs   int64                `storm:"index"`
	UserId     restApiV1.UserId     `storm:"index"`
	PlaylistId restApiV1.PlaylistId `storm:"index"`
}
