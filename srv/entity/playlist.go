package entity

import "lyra/restApiV1"

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
	PlaylistIdPk string
	SongIdPk     string
}

type PlaylistSongEntity struct {
	PlaylistSongId `storm:"id"`
	PlaylistId     string `storm:"index"`
	SongId         string `storm:"index"`
}

func NewPlaylistSongEntity(playlistId string, songId string) *PlaylistSongEntity {
	return &PlaylistSongEntity{
		PlaylistSongId: PlaylistSongId{
			PlaylistIdPk: playlistId,
			SongIdPk:     songId,
		},
		PlaylistId: playlistId,
		SongId:     songId,
	}
}

type OwnedUserPlaylistId struct {
	UserIdPk     string
	PlaylistIdPk string
}

type OwnedUserPlaylistEntity struct {
	OwnedUserPlaylistId `storm:"id"`
	UserId              string `storm:"index"`
	PlaylistId          string `storm:"index"`
}

func NewOwnedUserPlaylistEntity(userId string, playlistId string) *OwnedUserPlaylistEntity {
	return &OwnedUserPlaylistEntity{
		OwnedUserPlaylistId: OwnedUserPlaylistId{
			UserIdPk:     userId,
			PlaylistIdPk: playlistId,
		},
		UserId:     userId,
		PlaylistId: playlistId,
	}
}

type FavoritePlaylistId struct {
	UserIdPk     string
	PlaylistIdPk string
}

type FavoritePlaylistEntity struct {
	FavoritePlaylistId `storm:"id"`
	UserId             string `storm:"index"`
	PlaylistId         string `storm:"index"`
}

func NewFavoritePlaylistEntity(userId string, playlistId string) *FavoritePlaylistEntity {
	return &FavoritePlaylistEntity{
		FavoritePlaylistId: FavoritePlaylistId{
			UserIdPk:     userId,
			PlaylistIdPk: playlistId,
		},
		UserId:     userId,
		PlaylistId: playlistId,
	}
}
