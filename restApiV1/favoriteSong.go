package restApiV1

type FavoriteSongId struct {
	UserId UserId `json:"userId"`
	SongId SongId `json:"songId"`
}

type FavoriteSongMeta struct {
	Id FavoriteSongId `json:"id"`
}

type FavoriteSong struct {
	FavoriteSongMeta
	UpdateTs int64 `json:"updateTs"`
}
