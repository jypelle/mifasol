package entity

import "github.com/jypelle/mifasol/restApiV1"

// Song

type SongEntity struct {
	Id              restApiV1.SongId `storm:"id"`
	CreationTs      int64
	UpdateTs        int64  `storm:"index"`
	Name            string `storm:"index"`
	Format          restApiV1.SongFormat
	Size            int64
	BitDepth        restApiV1.SongBitDepth
	PublicationYear *int64
	AlbumId         restApiV1.AlbumId `storm:"index"`
	TrackNumber     *int64
	ArtistIds       []restApiV1.ArtistId
}

func (e *SongEntity) Fill(s *restApiV1.Song) {
	s.Id = e.Id
	s.CreationTs = e.CreationTs
	s.UpdateTs = e.UpdateTs
	s.Name = e.Name
	s.Format = e.Format
	s.Size = e.Size
	s.BitDepth = e.BitDepth
	s.PublicationYear = e.PublicationYear
	s.AlbumId = e.AlbumId
	s.TrackNumber = e.TrackNumber
	s.ArtistIds = e.ArtistIds
}

func (e *SongEntity) LoadMeta(s *restApiV1.SongMeta) {
	if s != nil {
		e.Name = s.Name
		e.Format = s.Format
		e.Size = s.Size
		e.BitDepth = s.BitDepth
		e.PublicationYear = s.PublicationYear
		e.AlbumId = s.AlbumId
		e.TrackNumber = s.TrackNumber
		e.ArtistIds = s.ArtistIds
	}
}

type DeletedSongEntity struct {
	Id       restApiV1.SongId `storm:"id"`
	DeleteTs int64            `storm:"index"`
}
