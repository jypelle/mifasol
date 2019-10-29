package restApiV1

// Album
const UnknownAlbumId AlbumId = "00000000000000000000000000"

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
