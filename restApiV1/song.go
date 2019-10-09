package restApiV1

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/vorbis"
	"io"
)

// Song

const (
	SongMimeTypeFlac = "audio/flac"
	SongMimeTypeOgg  = "audio/ogg"
	SongMimeTypeMp3  = "audio/mpeg"
)

type SongFormat int64

const (
	SongFormatUnknown SongFormat = iota
	SongFormatFlac
	SongFormatMp3
	SongFormatOgg
)

type SongBitDepth int64

const (
	SongBitDepthUnknown SongBitDepth = iota
	SongBitDepth16
	SongBitDepth24
)

func (s SongFormat) MimeType() string {
	switch s {
	case SongFormatFlac:
		return SongMimeTypeFlac
	case SongFormatOgg:
		return SongMimeTypeOgg
	case SongFormatMp3:
		return SongMimeTypeMp3
	}
	return "application/octet-stream"
}

func (s SongFormat) Extension() string {
	switch s {
	case SongFormatFlac:
		return ".flac"
	case SongFormatOgg:
		return ".ogg"
	case SongFormatMp3:
		return ".mp3"
	}
	return ".data"
}

func (s SongFormat) Decode() func(rc io.ReadCloser) (s beep.StreamSeekCloser, format beep.Format, err error) {
	switch s {
	case SongFormatFlac:
		return flac.Decode
	case SongFormatOgg:
		return vorbis.Decode
	case SongFormatMp3:
		return mp3.Decode
	}
	return nil
}

type Song struct {
	Id         string `json:"id" storm:"id"`
	CreationTs int64  `json:"creationTs"`
	UpdateTs   int64  `json:"updateTs" storm:"index"`
	SongMeta   `storm:"inline"`
}

type SongMeta struct {
	Name            string       `json:"name" storm:"index"`
	Format          SongFormat   `json:"format"`
	Size            int64        `json:"size"`
	BitDepth        SongBitDepth `json:"bitDepth"`
	PublicationYear *int64       `json:"publicationYear"`
	AlbumId         string       `json:"albumId" storm:"index"`
	TrackNumber     *int64       `json:"trackNumber"`
	ArtistIds       []string     `json:"artistIds"`
}

type SongNew struct {
	SongMeta `storm:"inline"`
	Content  []byte `json:"content"`
}

type DeletedSong struct {
	Id       string `json:"id" storm:"id"`
	DeleteTs int64  `json:"deleteTs" storm:"index"`
}
