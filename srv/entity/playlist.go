package entity

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

func (e *PlaylistEntity) LoadMeta(s *restApiV1.PlaylistMeta) {
	if s != nil {
		e.Name = s.Name
		e.SongIds = s.SongIds
		e.OwnerUserIds = s.OwnerUserIds
	}
}

type DeletedPlaylistEntity struct {
	Id       restApiV1.PlaylistId `storm:"id"`
	DeleteTs int64                `storm:"index"`
}

type PlaylistSongId struct {
	PlaylistId restApiV1.PlaylistId
	SongId     restApiV1.SongId
}

type PlaylistSongEntity struct {
	Id         string               `storm:"id"`
	PlaylistId restApiV1.PlaylistId `storm:"index"`
	SongId     restApiV1.SongId     `storm:"index"`
}

func NewPlaylistSongEntity(playlistId restApiV1.PlaylistId, songId restApiV1.SongId) *PlaylistSongEntity {
	return &PlaylistSongEntity{
		Id:         string(playlistId) + ":" + string(songId),
		PlaylistId: playlistId,
		SongId:     songId,
	}
}

type OwnedUserPlaylistEntity struct {
	Id         string               `storm:"id"`
	UserId     restApiV1.UserId     `storm:"index"`
	PlaylistId restApiV1.PlaylistId `storm:"index"`
}

func NewOwnedUserPlaylistEntity(userId restApiV1.UserId, playlistId restApiV1.PlaylistId) *OwnedUserPlaylistEntity {
	return &OwnedUserPlaylistEntity{
		Id:         string(userId) + string(playlistId),
		UserId:     userId,
		PlaylistId: playlistId,
	}
}
