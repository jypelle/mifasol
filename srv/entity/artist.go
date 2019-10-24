package entity

import "github.com/jypelle/mifasol/restApiV1"

// Artist

type ArtistEntity struct {
	Id         restApiV1.ArtistId `storm:"id"`
	CreationTs int64
	UpdateTs   int64  `storm:"index"`
	Name       string `storm:"index"`
}

func (e *ArtistEntity) Fill(s *restApiV1.Artist) {
	s.Id = e.Id
	s.CreationTs = e.CreationTs
	s.UpdateTs = e.UpdateTs
	s.Name = e.Name
}

func (e *ArtistEntity) LoadMeta(s *restApiV1.ArtistMeta) {
	if s != nil {
		e.Name = s.Name
	}
}

type DeletedArtistEntity struct {
	Id       restApiV1.ArtistId `storm:"id"`
	DeleteTs int64              `storm:"index"`
}

type ArtistSongId struct {
	ArtistId restApiV1.ArtistId
	SongId   string
}

type ArtistSongEntity struct {
	Id       string             `storm:"id"`
	ArtistId restApiV1.ArtistId `storm:"index"`
	SongId   restApiV1.SongId   `storm:"index"`
}

func NewArtistSongEntity(artistId restApiV1.ArtistId, songId restApiV1.SongId) *ArtistSongEntity {
	return &ArtistSongEntity{
		Id:       string(artistId) + ":" + string(songId),
		ArtistId: artistId,
		SongId:   songId,
	}
}
