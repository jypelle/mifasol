package entity

import "mifasol/restApiV1"

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
	Id     PlaylistSongId `storm:"id"`
	SongId string         `storm:"index"`
}

func NewPlaylistSongEntity(playlistId string, songId string) *PlaylistSongEntity {
	return &PlaylistSongEntity{
		Id: PlaylistSongId{
			PlaylistId: playlistId,
			SongId:     songId,
		},
		SongId: songId,
	}
}

type OwnedUserPlaylistId struct {
	UserId     string
	PlaylistId string
}

type OwnedUserPlaylistEntity struct {
	Id     OwnedUserPlaylistId `storm:"id"`
	UserId string              `storm:"index"`
}

func NewOwnedUserPlaylistEntity(userId string, playlistId string) *OwnedUserPlaylistEntity {
	return &OwnedUserPlaylistEntity{
		Id: OwnedUserPlaylistId{
			UserId:     userId,
			PlaylistId: playlistId,
		},
		UserId: userId,
	}
}

type FavoritePlaylistId struct {
	UserId     string
	PlaylistId string
}

type FavoritePlaylistEntity struct {
	Id         FavoritePlaylistId `storm:"id"`
	UserId     string             `storm:"index"`
	PlaylistId string             `storm:"index"`
}

func NewFavoritePlaylistEntity(userId string, playlistId string) *FavoritePlaylistEntity {
	return &FavoritePlaylistEntity{
		Id: FavoritePlaylistId{
			UserId:     userId,
			PlaylistId: playlistId,
		},
		UserId:     userId,
		PlaylistId: playlistId,
	}
}
