package restApiV1

// Playlist

type Playlist struct {
	Id              string `json:"id" storm:"id"`
	CreationTs      int64  `json:"creationTs"`
	UpdateTs        int64  `json:"updateTs" storm:"index"`
	ContentUpdateTs int64  `json:"contentUpdateTs" storm:"index"`
	PlaylistMeta    `storm:"inline"`
}

type PlaylistMeta struct {
	Name         string   `json:"name" storm:"index"`
	SongIds      []string `json:"songIds"`
	OwnerUserIds []string `json:"ownerUserIds"`
}

type DeletedPlaylist struct {
	Id       string `json:"id" storm:"id"`
	DeleteTs int64  `json:"deleteTs" storm:"index"`
}

type PlaylistSong struct {
	PlaylistSongId `storm:"id"`
}

type PlaylistSongId struct {
	PlaylistId string `json:"playlistId" storm:"index"`
	SongId     string `json:"songId" storm:"index"`
}

type OwnedUserPlaylist struct {
	OwnedUserPlaylistId `storm:"id"`
}

type OwnedUserPlaylistId struct {
	UserId     string `json:"userId" storm:"index"`
	PlaylistId string `json:"playlistId" storm:"index"`
}

type FavoritePlaylist struct {
	FavoritePlaylistId `storm:"id"`
}

type FavoritePlaylistId struct {
	UserId     string `json:"userId" storm:"index"`
	PlaylistId string `json:"playlistId" storm:"index"`
}
