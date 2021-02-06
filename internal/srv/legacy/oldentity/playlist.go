package oldentity

import (
	"github.com/jypelle/mifasol/restApiV1"
)

// Playlist

type PlaylistEntity struct {
	Id              restApiV1.PlaylistId `storm:"id"`
	CreationTs      int64
	UpdateTs        int64  `storm:"index"`
	ContentUpdateTs int64  `storm:"index"`
	Name            string `storm:"index"`
	SongIds         []restApiV1.SongId
	OwnerUserIds    []restApiV1.UserId
}

func (e *PlaylistEntity) Fill(s *restApiV1.Playlist) {
	s.Id = e.Id
	s.CreationTs = e.CreationTs
	s.UpdateTs = e.UpdateTs
	s.ContentUpdateTs = e.ContentUpdateTs
	s.Name = e.Name
	s.SongIds = e.SongIds
	s.OwnerUserIds = e.OwnerUserIds
}

type DeletedPlaylistEntity struct {
	Id       restApiV1.PlaylistId `storm:"id"`
	DeleteTs int64                `storm:"index"`
}
