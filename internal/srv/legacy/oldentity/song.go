package oldentity

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
	ExplicitFg      bool
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
	s.ExplicitFg = e.ExplicitFg
	s.ArtistIds = e.ArtistIds
}

type DeletedSongEntity struct {
	Id       restApiV1.SongId `storm:"id"`
	DeleteTs int64            `storm:"index"`
}
