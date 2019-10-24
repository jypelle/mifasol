package restApiV1

type SyncReport struct {
	Songs                      []Song               `json:"songs"`
	DeletedSongIds             []SongId             `json:"deletedSongIds"`
	Artists                    []Artist             `json:"artists"`
	DeletedArtistIds           []ArtistId           `json:"deletedArtistIds"`
	Albums                     []Album              `json:"albums"`
	DeletedAlbumIds            []AlbumId            `json:"deletedAlbumIds"`
	Playlists                  []Playlist           `json:"playlists"`
	DeletedPlaylistIds         []PlaylistId         `json:"deletedPlaylistIds"`
	Users                      []User               `json:"users"`
	DeletedUserIds             []UserId             `json:"deletedUserIds"`
	FavoritePlaylists          []FavoritePlaylist   `json:"favoritePlaylists"`
	DeletedFavoritePlaylistIds []FavoritePlaylistId `json:"deletedFavoritePlaylistIds"`
	SyncTs                     int64                `json:"syncTs"`
}

type FileSyncSong struct {
	Id       SongId `json:"id"`
	UpdateTs int64  `json:"updateTs"`
	Filepath string `json:"filepath"`
}

type FileSyncReport struct {
	FileSyncSongs      []FileSyncSong `json:"fileSyncSongs"`
	DeletedSongIds     []SongId       `json:"deletedSongIds"`
	Playlists          []Playlist     `json:"playlists"`
	DeletedPlaylistIds []PlaylistId   `json:"deletedPlaylistIds"`
	SyncTs             int64          `json:"syncTs"`
}
