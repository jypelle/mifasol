package entity

import (
	"database/sql"
	"github.com/jypelle/mifasol/restApiV1"
)

// Song

type SongEntity struct {
	SongId          restApiV1.SongId       `db:"song_id"`
	CreationTs      int64                  `db:"creation_ts"`
	UpdateTs        int64                  `db:"update_ts"`
	Name            string                 `db:"name"`
	Format          restApiV1.SongFormat   `db:"format"`
	Size            int64                  `db:"size"`
	BitDepth        restApiV1.SongBitDepth `db:"bit_depth"`
	PublicationYear sql.NullInt64          `db:"publication_year"`
	AlbumId         restApiV1.AlbumId      `db:"album_id"`
	TrackNumber     sql.NullInt64          `db:"track_number"`
	ExplicitFg      bool                   `db:"explicit_fg"`
}

func (e *SongEntity) Fill(s *restApiV1.Song) {
	s.Id = e.SongId
	s.CreationTs = e.CreationTs
	s.UpdateTs = e.UpdateTs
	s.Name = e.Name
	s.Format = e.Format
	s.Size = e.Size
	s.BitDepth = e.BitDepth
	if e.PublicationYear.Valid {
		s.PublicationYear = &e.PublicationYear.Int64
	} else {
		s.PublicationYear = nil
	}
	s.AlbumId = e.AlbumId
	if e.TrackNumber.Valid {
		s.TrackNumber = &e.TrackNumber.Int64
	} else {
		s.TrackNumber = nil
	}
	s.ExplicitFg = e.ExplicitFg
}

func (e *SongEntity) LoadMeta(s *restApiV1.SongMeta) {
	if s != nil {
		e.Name = s.Name
		e.Format = s.Format
		e.Size = s.Size
		e.BitDepth = s.BitDepth
		if s.PublicationYear != nil {
			e.PublicationYear.Int64 = *s.PublicationYear
			e.PublicationYear.Valid = true
		} else {
			e.PublicationYear.Valid = false
		}
		e.AlbumId = s.AlbumId
		if s.TrackNumber != nil {
			e.TrackNumber.Int64 = *s.TrackNumber
			e.TrackNumber.Valid = true
		} else {
			e.TrackNumber.Valid = false
		}
		e.ExplicitFg = s.ExplicitFg
	}
}

type DeletedSongEntity struct {
	SongId   restApiV1.SongId `db:"song_id"`
	DeleteTs int64            `db:"delete_ts"`
}
