package entity

import "github.com/jypelle/mifasol/restApiV1"

// Artist

type ArtistEntity struct {
	ArtistId   restApiV1.ArtistId `db:"artist_id"`
	CreationTs int64              `db:"creation_ts"`
	UpdateTs   int64              `db:"update_ts"`
	Name       string             `db:"name"`
}

func (e *ArtistEntity) Fill(s *restApiV1.Artist) {
	s.Id = e.ArtistId
	s.CreationTs = e.CreationTs
	s.UpdateTs = e.UpdateTs
	s.Name = e.Name
}

func (e *ArtistEntity) LoadMeta(s *restApiV1.ArtistMeta) {
	if s != nil {
		e.Name = s.Name
	}
}

type ArtistSongEntity struct {
	ArtistId restApiV1.ArtistId `db:"artist_id"`
	SongId   restApiV1.SongId   `db:"song_id"`
}

type DeletedArtistEntity struct {
	ArtistId restApiV1.ArtistId `db:"artist_id"`
	DeleteTs int64              `db:"delete_ts"`
}
