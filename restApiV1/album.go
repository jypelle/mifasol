package restApiV1

// Album

type Album struct {
	Id         string   `json:"id"`
	CreationTs int64    `json:"creationTs"`
	UpdateTs   int64    `json:"updateTs"`
	ArtistIds  []string `json:"artistIds"`
	AlbumMeta
}

type AlbumMeta struct {
	Name string `json:"name"`
}
