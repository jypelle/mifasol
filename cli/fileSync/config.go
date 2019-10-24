package fileSync

import "github.com/jypelle/mifasol/restApiV1"

const FileSyncFilename = ".mifasolFileSync.json"

type FileSyncConfig struct {
	LastFileSyncTs         int64                                           `json:"lastFileSyncTs"`
	FileSyncLocalSongs     map[restApiV1.SongId]*FileSyncLocalSong         `json:"localSongs"`
	FileSyncLocalPlaylists map[restApiV1.PlaylistId]*FileSyncLocalPlaylist `json:"localPlaylists"`
}

type FileSyncLocalSong struct {
	UpdateTs int64  `json:"updateTs"`
	Filepath string `json:"filepath"`
}

type FileSyncLocalPlaylist struct {
	UpdateTs int64  `json:"updateTs"`
	Filepath string `json:"filepath"`
}
