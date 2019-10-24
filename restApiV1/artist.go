package restApiV1

// Artist

type Artist struct {
	Id         string `json:"id"`
	CreationTs int64  `json:"creationTs"`
	UpdateTs   int64  `json:"updateTs"`
	ArtistMeta
}

type ArtistMeta struct {
	Name string `json:"name"`
}
