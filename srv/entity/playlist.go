package entity

import (
	"github.com/jypelle/mifasol/restApiV1"
)

// Playlist

type PlaylistEntity struct {
	Id              string `storm:"id"`
	CreationTs      int64
	UpdateTs        int64  `storm:"index"`
	ContentUpdateTs int64  `storm:"index"`
	Name            string `storm:"index"`
	SongIds         []string
	OwnerUserIds    []string
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
	Id       string `storm:"id"`
	DeleteTs int64  `storm:"index"`
}

type PlaylistSongId struct {
	PlaylistId string
	SongId     string
}

type PlaylistSongEntity struct {
	Id         string `storm:"id"`
	PlaylistId string `storm:"index"`
	SongId     string `storm:"index"`
}

func NewPlaylistSongEntity(playlistId string, songId string) *PlaylistSongEntity {
	return &PlaylistSongEntity{
		Id:         playlistId + ":" + songId,
		PlaylistId: playlistId,
		SongId:     songId,
	}
}

type OwnedUserPlaylistEntity struct {
	Id         string `storm:"id"`
	UserId     string `storm:"index"`
	PlaylistId string `storm:"index"`
}

func NewOwnedUserPlaylistEntity(userId string, playlistId string) *OwnedUserPlaylistEntity {
	return &OwnedUserPlaylistEntity{
		Id:         userId + playlistId,
		UserId:     userId,
		PlaylistId: playlistId,
	}
}
