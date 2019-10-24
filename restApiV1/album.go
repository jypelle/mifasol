package restApiV1

// Album
type AlbumId string

type Album struct {
	Id         AlbumId    `json:"id"`
	CreationTs int64      `json:"creationTs"`
	UpdateTs   int64      `json:"updateTs"`
	ArtistIds  []ArtistId `json:"artistIds"`
	AlbumMeta
}

type AlbumMeta struct {
	Name string `json:"name"`
}
