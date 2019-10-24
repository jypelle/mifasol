package restApiV1

// Playlist

const IncomingPlaylistId = "00000000000000000000000000"

type PlaylistId string

type Playlist struct {
	Id              PlaylistId `json:"id"`
	CreationTs      int64      `json:"creationTs"`
	UpdateTs        int64      `json:"updateTs"`
	ContentUpdateTs int64      `json:"contentUpdateTs"`
	PlaylistMeta
}

type PlaylistMeta struct {
	Name         string   `json:"name"`
	SongIds      []SongId `json:"songIds"`
	OwnerUserIds []UserId `json:"ownerUserIds"`
}
