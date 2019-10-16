package restApiV1

// Playlist

const IncomingPlaylistId = "00000000000000000000000000"

type Playlist struct {
	Id              string `json:"id"`
	CreationTs      int64  `json:"creationTs"`
	UpdateTs        int64  `json:"updateTs"`
	ContentUpdateTs int64  `json:"contentUpdateTs"`
	PlaylistMeta
}

type PlaylistMeta struct {
	Name         string   `json:"name"`
	SongIds      []string `json:"songIds"`
	OwnerUserIds []string `json:"ownerUserIds"`
}
