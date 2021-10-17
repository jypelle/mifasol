package restApiV1

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

func (s *SongMeta) Copy() *SongMeta {
	var newSongMeta = *s
	if s.PublicationYear != nil {
		newPublicationYear := *s.PublicationYear
		newSongMeta.PublicationYear = &newPublicationYear
	}
	if s.TrackNumber != nil {
		newTrackNumber := *s.TrackNumber
		newSongMeta.TrackNumber = &newTrackNumber
	}
	newSongMeta.ArtistIds = make([]ArtistId, len(s.ArtistIds))
	copy(newSongMeta.ArtistIds, s.ArtistIds)
	return &newSongMeta
}

type SongNew struct {
	SongMeta
	Content []byte `json:"content"`
}
