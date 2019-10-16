package restApiV1

type FavoritePlaylistId struct {
	UserId     string `json:"userId"`
	PlaylistId string `json:"playlistId"`
}

type FavoritePlaylistMeta struct {
	Id FavoritePlaylistId `json:"id"`
}

type FavoritePlaylist struct {
	FavoritePlaylistMeta
	UpdateTs int64 `json:"updateTs"`
}
