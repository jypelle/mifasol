package fileSync

const FileSyncFilename = ".mifasolFileSync.json"

type FileSyncConfig struct {
	LastFileSyncTs         int64                             `json:"lastFileSyncTs"`
	FileSyncLocalSongs     map[string]*FileSyncLocalSong     `json:"localSongs"`
	FileSyncLocalPlaylists map[string]*FileSyncLocalPlaylist `json:"localPlaylists"`
}

type FileSyncLocalSong struct {
	UpdateTs int64  `json:"updateTs"`
	Filepath string `json:"filepath"`
}

type FileSyncLocalPlaylist struct {
	UpdateTs int64  `json:"updateTs"`
	Filepath string `json:"filepath"`
}
