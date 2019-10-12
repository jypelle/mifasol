package restApiV1

// Artist

type Artist struct {
	Id         string `json:"id" storm:"id"`
	CreationTs int64  `json:"creationTs"`
	UpdateTs   int64  `json:"updateTs" storm:"index"`
	ArtistMeta `storm:"inline"`
}

type ArtistMeta struct {
	Name string `json:"name" storm:"index"`
}

type DeletedArtist struct {
	Id       string `json:"id" storm:"id"`
	DeleteTs int64  `json:"deleteTs" storm:"index"`
}

type ArtistSongId struct {
	ArtistIdPk string
	SongIdPk   string
}

type ArtistSong struct {
	ArtistSongId `storm:"id"`
	ArtistId     string `json:"artistId" storm:"index"`
	SongId       string `json:"songId" storm:"index"`
}
