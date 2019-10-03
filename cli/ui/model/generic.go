package model

type IdType int64

const (
	IdTypeUnknown IdType = iota
	IdTypeArtist
	IdTypeAlbum
	IdTypeSong
	IdTypePlaylist
	IdTypeUser
)
