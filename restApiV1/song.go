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

func (s SongBitDepth) String() string {
	switch s {
	case SongBitDepth16:
		return "16b"
	case SongBitDepth24:
		return "24b"
	}
	return "unknown"
}

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

func (s SongFormat) String() string {
	switch s {
	case SongFormatFlac:
		return "flac"
	case SongFormatOgg:
		return "ogg"
	case SongFormatMp3:
		return "mp3"
	}
	return "data"
}

func (s SongFormat) Decode() func(rc io.ReadCloser) (s beep.StreamSeekCloser, format beep.Format, err error) {
	switch s {
	case SongFormatFlac:
		return func(rc io.ReadCloser) (s beep.StreamSeekCloser, format beep.Format, err error) {
			return flac.Decode(rc)
		}
	case SongFormatOgg:
		return vorbis.Decode
	case SongFormatMp3:
		return mp3.Decode
	}
	return nil
}

type SongId string

type Song struct {
	Id         SongId `json:"id"`
	CreationTs int64  `json:"creationTs"`
	UpdateTs   int64  `json:"updateTs"`
	SongMeta
}

type SongMeta struct {
	Name            string       `json:"name"`
	Format          SongFormat   `json:"format"`
	Size            int64        `json:"size"`
	BitDepth        SongBitDepth `json:"bitDepth"`
	PublicationYear *int64       `json:"publicationYear"`
	AlbumId         AlbumId      `json:"albumId"`
	TrackNumber     *int64       `json:"trackNumber"`
	ArtistIds       []ArtistId   `json:"artistIds"`
	ExplicitFg      bool         `json:"explicitFg"`
}

type SongNew struct {
	SongMeta
	Content []byte `json:"content"`
}
