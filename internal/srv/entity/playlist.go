package entity

import (
	"github.com/jypelle/mifasol/restApiV1"
)

// Playlist

type PlaylistEntity struct {
	PlaylistId      restApiV1.PlaylistId `db:"playlist_id"`
	CreationTs      int64                `db:"creation_ts"`
	UpdateTs        int64                `db:"update_ts"`
	ContentUpdateTs int64                `db:"content_update_ts"`
	Name            string               `db:"name"`
}

func (e *PlaylistEntity) Fill(s *restApiV1.Playlist) {
	s.Id = e.PlaylistId
	s.CreationTs = e.CreationTs
	s.UpdateTs = e.UpdateTs
	s.ContentUpdateTs = e.ContentUpdateTs
	s.Name = e.Name
}

func (e *PlaylistEntity) LoadMeta(s *restApiV1.PlaylistMeta) {
	if s != nil {
		e.Name = s.Name
	}
}

type PlaylistSongEntity struct {
	PlaylistId restApiV1.PlaylistId `db:"playlist_id"`
	Position   int64                `db:"position"`
	SongId     restApiV1.SongId     `db:"song_id"`
}

func NewPlaylistSongEntity(playlistId restApiV1.PlaylistId, position int64, songId restApiV1.SongId) *PlaylistSongEntity {
	return &PlaylistSongEntity{
		PlaylistId: playlistId,
		Position:   position,
		SongId:     songId,
	}
}

type PlaylistOwnedUserEntity struct {
	PlaylistId restApiV1.PlaylistId `db:"playlist_id"`
	UserId     restApiV1.UserId     `db:"user_id"`
}

func NewPlaylistOwnedUserEntity(userId restApiV1.UserId, playlistId restApiV1.PlaylistId) *PlaylistOwnedUserEntity {
	return &PlaylistOwnedUserEntity{
		PlaylistId: playlistId,
		UserId:     userId,
	}
}

type DeletedPlaylistEntity struct {
	PlaylistId restApiV1.PlaylistId `db:"playlist_id"`
	DeleteTs   int64                `db:"delete_ts"`
}
