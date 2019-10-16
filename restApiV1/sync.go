package restApiV1

type SyncReport struct {
	Songs                      []Song               `json:"songs"`
	DeletedSongIds             []string             `json:"deletedSongIds"`
	Artists                    []Artist             `json:"artists"`
	DeletedArtistIds           []string             `json:"deletedArtistIds"`
	Albums                     []Album              `json:"albums"`
	DeletedAlbumIds            []string             `json:"deletedAlbumIds"`
	Playlists                  []Playlist           `json:"playlists"`
	DeletedPlaylistIds         []string             `json:"deletedPlaylistIds"`
	Users                      []User               `json:"users"`
	DeletedUserIds             []string             `json:"deletedUserIds"`
	FavoritePlaylists          []FavoritePlaylist   `json:"favoritePlaylists"`
	DeletedFavoritePlaylistIds []FavoritePlaylistId `json:"deletedFavoritePlaylistIds"`
	SyncTs                     int64                `json:"syncTs"`
}

type FileSyncSong struct {
	Id       string `json:"id"`
	UpdateTs int64  `json:"updateTs"`
	Filepath string `json:"filepath"`
}

type FileSyncReport struct {
	FileSyncSongs      []FileSyncSong `json:"fileSyncSongs"`
	DeletedSongIds     []string       `json:"deletedSongIds"`
	Playlists          []Playlist     `json:"playlists"`
	DeletedPlaylistIds []string       `json:"deletedPlaylistIds"`
	SyncTs             int64          `json:"syncTs"`
}
