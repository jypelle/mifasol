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

type DeletedPlaylist struct {
	Id       string `json:"id" storm:"id"`
	DeleteTs int64  `json:"deleteTs" storm:"index"`
}

type PlaylistSongId struct {
	PlaylistIdPk string
	SongIdPk     string
}

type PlaylistSong struct {
	PlaylistSongId `storm:"id"`
	PlaylistId     string `json:"playlistId" storm:"index"`
	SongId         string `json:"songId" storm:"index"`
}

func NewPlaylistSong(playlistId string, songId string) *PlaylistSong {
	return &PlaylistSong{
		PlaylistSongId: PlaylistSongId{
			PlaylistIdPk: playlistId,
			SongIdPk:     songId,
		},
		PlaylistId: playlistId,
		SongId:     songId,
	}
}

type OwnedUserPlaylistId struct {
	UserIdPk     string
	PlaylistIdPk string
}

type OwnedUserPlaylist struct {
	OwnedUserPlaylistId `storm:"id"`
	UserId              string `json:"userId" storm:"index"`
	PlaylistId          string `json:"playlistId" storm:"index"`
}

func NewOwnedUserPlaylist(userId string, playlistId string) *OwnedUserPlaylist {
	return &OwnedUserPlaylist{
		OwnedUserPlaylistId: OwnedUserPlaylistId{
			UserIdPk:     userId,
			PlaylistIdPk: playlistId,
		},
		UserId:     userId,
		PlaylistId: playlistId,
	}
}

type FavoritePlaylistId struct {
	UserIdPk     string
	PlaylistIdPk string
}

type FavoritePlaylist struct {
	FavoritePlaylistId `storm:"id"`
	UserId             string `json:"userId" storm:"index"`
	PlaylistId         string `json:"playlistId" storm:"index"`
}

func NewFavoritePlaylist(userId string, playlistId string) *FavoritePlaylist {
	return &FavoritePlaylist{
		FavoritePlaylistId: FavoritePlaylistId{
			UserIdPk:     userId,
			PlaylistIdPk: playlistId,
		},
		UserId:     userId,
		PlaylistId: playlistId,
	}
}
