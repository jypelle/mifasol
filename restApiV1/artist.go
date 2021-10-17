package restApiV1

// Artist

var UnknownArtistId ArtistId = "00000000000000000000000000"

type ArtistId string

type Artist struct {
	Id         ArtistId `json:"id"`
	CreationTs int64    `json:"creationTs"`
	UpdateTs   int64    `json:"updateTs"`
	ArtistMeta
}

type ArtistMeta struct {
	Name string `json:"name"`
}

func (a *ArtistMeta) Copy() *ArtistMeta {
	var newArtistMeta = *a
	return &newArtistMeta
}
