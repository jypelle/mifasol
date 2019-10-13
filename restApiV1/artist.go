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
