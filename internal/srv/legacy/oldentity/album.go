package oldentity

import "github.com/jypelle/mifasol/restApiV1"

// Album

type AlbumEntity struct {
	Id         restApiV1.AlbumId `storm:"id"`
	CreationTs int64
	UpdateTs   int64  `storm:"index"`
	Name       string `storm:"index"`
	ArtistIds  []restApiV1.ArtistId
}

func (e *AlbumEntity) Fill(s *restApiV1.Album) {
	s.Id = e.Id
	s.CreationTs = e.CreationTs
	s.UpdateTs = e.UpdateTs
	s.Name = e.Name
	s.ArtistIds = e.ArtistIds
}

type DeletedAlbumEntity struct {
	Id       restApiV1.AlbumId `storm:"id"`
	DeleteTs int64             `storm:"index"`
}
