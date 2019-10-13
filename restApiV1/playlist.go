package restApiV1

// Playlist

const IncomingPlaylistId = "00000000000000000000000000"

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
