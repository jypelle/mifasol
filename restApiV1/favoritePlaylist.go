package restApiV1

type FavoritePlaylistId struct {
	UserId     UserId     `json:"userId"`
	PlaylistId PlaylistId `json:"playlistId"`
}

type FavoritePlaylistMeta struct {
	Id FavoritePlaylistId `json:"id"`
}

type FavoritePlaylist struct {
	FavoritePlaylistMeta
	UpdateTs int64 `json:"updateTs"`
}
