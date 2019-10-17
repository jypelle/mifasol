package entity

import "github.com/jypelle/mifasol/restApiV1"

// Album

type AlbumEntity struct {
	Id         string `storm:"id"`
	CreationTs int64
	UpdateTs   int64  `storm:"index"`
	Name       string `storm:"index"`
	ArtistIds  []string
}

func (e *AlbumEntity) Fill(s *restApiV1.Album) {
	s.Id = e.Id
	s.CreationTs = e.CreationTs
	s.UpdateTs = e.UpdateTs
	s.Name = e.Name
	s.ArtistIds = e.ArtistIds
}

func (e *AlbumEntity) LoadMeta(s *restApiV1.AlbumMeta) {
	if s != nil {
		e.Name = s.Name
	}
}

type DeletedAlbumEntity struct {
	Id       string `storm:"id"`
	DeleteTs int64  `storm:"index"`
}
