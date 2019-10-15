package entity

import "mifasol/restApiV1"

// Artist

type ArtistEntity struct {
	Id         string `storm:"id"`
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
	Id       string `storm:"id"`
	DeleteTs int64  `storm:"index"`
}

type ArtistSongId struct {
	ArtistId string
	SongId   string
}

type ArtistSongEntity struct {
	Id       ArtistSongId `storm:"id"`
	ArtistId string       `storm:"index"`
	SongId   string       `storm:"index"`
}

func NewArtistSongEntity(artistId string, songId string) *ArtistSongEntity {
	return &ArtistSongEntity{
		Id: ArtistSongId{
			ArtistId: artistId,
			SongId:   songId,
		},
		ArtistId: artistId,
		SongId:   songId,
	}
}
