package entity

import "github.com/jypelle/mifasol/restApiV1"

// Album

type AlbumEntity struct {
	AlbumId    restApiV1.AlbumId `db:"album_id"`
	CreationTs int64             `db:"creation_ts"`
	UpdateTs   int64             `db:"update_ts"`
	Name       string            `db:"name"`
}

func (e *AlbumEntity) Fill(s *restApiV1.Album) {
	s.Id = e.AlbumId
	s.CreationTs = e.CreationTs
	s.UpdateTs = e.UpdateTs
	s.Name = e.Name
}

func (e *AlbumEntity) LoadMeta(s *restApiV1.AlbumMeta) {
	if s != nil {
		e.Name = s.Name
	}
}

type DeletedAlbumEntity struct {
	AlbumId  restApiV1.AlbumId `db:"album_id"`
	DeleteTs int64             `db:"delete_ts"`
}
