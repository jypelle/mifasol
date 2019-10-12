package restApiV1

// Album

type Album struct {
	Id         string   `json:"id" storm:"id"`
	CreationTs int64    `json:"creationTs"`
	UpdateTs   int64    `json:"updateTs" storm:"index"`
	ArtistIds  []string `json:"artistIds"`
	AlbumMeta  `storm:"inline"`
}

type AlbumMeta struct {
	Name string `json:"name" storm:"index"`
}

type DeletedAlbum struct {
	Id       string `json:"id" storm:"id"`
	DeleteTs int64  `json:"deleteTs" storm:"index"`
}
