package fileSync

import (
	"encoding/json"
	"fmt"
	"github.com/jypelle/mifasol/internal/cli/config"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/sirupsen/logrus"
	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type App struct {
	config.ClientConfig
	restClient *restClientV1.RestClient

	fileSyncMusicFolder     string
	fileSyncConfig          FileSyncConfig
	doneChannel             chan bool
	interruptChannel        chan os.Signal
	interruptRequestChannel chan bool
}

func NewApp(clientConfig config.ClientConfig, restClient *restClientV1.RestClient, fileSyncMusicFolder string) *App {
	app := &App{
		ClientConfig:        clientConfig,
		restClient:          restClient,
		fileSyncMusicFolder: strings.Replace(fileSyncMusicFolder, "\\", "/", -1),
		fileSyncConfig: FileSyncConfig{
			LastFileSyncTs:         0,
			FileSyncLocalSongs:     make(map[restApiV1.SongId]*FileSyncLocalSong),
			FileSyncLocalPlaylists: make(map[restApiV1.PlaylistId]*FileSyncLocalPlaylist),
		},
		doneChannel:             make(chan bool),
		interruptChannel:        make(chan os.Signal),
		interruptRequestChannel: make(chan bool),
	}

	return app
}

func (a *App) Sync() {
	signal.Notify(a.interruptChannel, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)

	go a.sync()

	select {
	case <-a.doneChannel:
		return
	case <-a.interruptChannel:
		fmt.Println("Interruption requested: finishing current file transfert before exiting\n")
		a.interruptRequestChannel <- true
		<-a.doneChannel
	}

}

func (a *App) sync() {
	defer func() { a.doneChannel <- true }()
	synchroAborded := false

	// Check music folder
	_, err := os.Stat(a.fileSyncMusicFolder)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Fatalf("Folder doesn't exist: %s", a.fileSyncMusicFolder)
		} else {
			logrus.Fatalf("Unable to access music folder: %s", a.fileSyncMusicFolder)
		}
	}

	// Check configuration file
	_, err = os.Stat(a.getCompleteFileSyncFilename())
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Fatalf("Folder has not been initialized")
		} else {
			logrus.Fatalf("Unable to access: %s", a.getCompleteFileSyncFilename())
		}
	}

	// Open configuration file
	rawConfig, err := ioutil.ReadFile(a.getCompleteFileSyncFilename())
	if err != nil {
		logrus.Fatalf("Unable to load: %v\n", err)
	}

	// Interpret configuration file
	err = json.Unmarshal(rawConfig, &a.fileSyncConfig)
	if err != nil {
		logrus.Fatalf("Unable to interpret: %v\n", err)
	}

	// Read file sync report
	fileSyncReport, cliErr := a.restClient.ReadFileSyncReport(a.fileSyncConfig.LastFileSyncTs, a.restClient.UserId())
	if cliErr != nil {
		logrus.Fatalf("Unable to retrieve songs data: %v\n", cliErr)
	}

	progressContainer := mpb.New(mpb.WithWidth(50))

	fmt.Println("Start synchronization")
	songSyncErrors := 0
	songSyncDeletedSongs := 0
	songSyncUpdatedSongs := 0
	playlistSyncErrors := 0
	playlistSyncDeletedPlaylists := 0
	playlistSyncUpdatedPlaylists := 0

	// Sync song files

	songsBar := progressContainer.AddBar(int64(len(fileSyncReport.FileSyncSongs)+1),
		mpb.PrependDecorators(
			// simple name decorator
			decor.Name("Sync songs"),
			decor.Percentage(decor.WCSyncSpace),
		),
	)

	// Delete songs

	for _, songId := range fileSyncReport.DeletedSongIds {
		if synchroAborded {
			break
		}
		func() {
			defer func() {
				select {
				case <-a.interruptRequestChannel:
					synchroAborded = true
					songsBar.Abort(false)
				default:
				}
			}()

			if fileSyncLocalSong, ok := a.fileSyncConfig.FileSyncLocalSongs[songId]; ok {
				fullpath := a.fileSyncMusicFolder + "/songs/" + fileSyncLocalSong.Filepath
				err := os.Remove(fullpath)
				if err != nil && err != syscall.ENOENT {
					logrus.Warningf("Unable to delete \"%s\": %s\n", fullpath, err.Error())
					songSyncErrors++
				} else {
					delete(a.fileSyncConfig.FileSyncLocalSongs, songId)
					songSyncDeletedSongs++
					deleteVoidParentFolder(fullpath)
					logrus.Debugf("\"%s\" deleted\n", fileSyncLocalSong.Filepath)
				}
			}
		}()

	}

	// Update songs
	if !synchroAborded {
		songsBar.Increment()

		for _, fileSyncSong := range fileSyncReport.FileSyncSongs {
			if synchroAborded {
				break
			}
			func() {
				defer func() {
					songsBar.Increment()
					select {
					case <-a.interruptRequestChannel:
						synchroAborded = true
						songsBar.Abort(false)
					default:
					}
				}()

				// Continue if already updated
				if fileSyncLocalSong, ok := a.fileSyncConfig.FileSyncLocalSongs[fileSyncSong.Id]; ok {
					if fileSyncSong.UpdateTs <= fileSyncLocalSong.UpdateTs {
						return
					}
				}

				var err error
				// Read song content
				reader, contentLength, apiErr := a.restClient.ReadSongContent(fileSyncSong.Id)
				if apiErr != nil {
					logrus.Warningf("Unable to read \"%s\" from mifasolsrv: %v\n", fileSyncSong.Filepath, apiErr)
					songSyncErrors++
					return
				}
				defer reader.Close()

				indexLastSeparator := strings.LastIndex(fileSyncSong.Filepath, "/")
				truncatedFilePath := tool.CharacterTruncate(fileSyncSong.Filepath[indexLastSeparator+1:], 23)

				songBar := progressContainer.AddBar(contentLength,
					mpb.PrependDecorators(
						decor.Name("Download  "),
						decor.Percentage(decor.WCSyncSpace),
					),
					mpb.AppendDecorators(
						decor.CountersKibiByte("%6.1f / %6.1f"),
						decor.Name(" "+truncatedFilePath),
					),
					mpb.BarRemoveOnComplete(),
				)

				proxyReader := songBar.ProxyReader(reader)

				fullNewpath := a.fileSyncMusicFolder + "/songs/" + fileSyncSong.Filepath

				// Create song folder(s)
				index := strings.LastIndex(fullNewpath, "/")
				if index != -1 {
					err = os.MkdirAll(fullNewpath[:index], 0770)
					if err != nil {
						logrus.Warningf("Unable to create folders for \"%s\": %s\n", fileSyncSong.Filepath, err.Error())
						songSyncErrors++
						return
					}
				}

				// Write song content
				file, err := os.Create(fullNewpath + ".tmp")
				if err != nil {
					logrus.Warningf("Unable to create file \"%s\": %s\n", fullNewpath+".tmp", err.Error())
					songSyncErrors++
					return
				}

				_, err = io.Copy(file, proxyReader)
				file.Close()

				if err != nil {
					logrus.Warningf("Unable to write file \"%s\": %s\n", fullNewpath+".tmp", err.Error())
					songSyncErrors++
					return
				}

				// Delete old song
				if fileSyncLocalSong, ok := a.fileSyncConfig.FileSyncLocalSongs[fileSyncSong.Id]; ok {
					oldPath := a.fileSyncMusicFolder + "/songs/" + fileSyncLocalSong.Filepath
					os.Remove(oldPath)
					deleteVoidParentFolder(oldPath)

				}

				// Rename new song
				err = os.Rename(fullNewpath+".tmp", fullNewpath)
				if err != nil {
					logrus.Warningf("Unable to rename file \"%s\": %s\n", fullNewpath+".tmp", err.Error())
					songSyncErrors++
					return
				}

				// Update song sync data
				a.fileSyncConfig.FileSyncLocalSongs[fileSyncSong.Id] = &FileSyncLocalSong{
					Filepath: fileSyncSong.Filepath,
					UpdateTs: fileSyncSong.UpdateTs,
				}
				songSyncUpdatedSongs++
			}()
		}
	}

	// Sync playlist files

	if !synchroAborded {

		playlistsBar := progressContainer.AddBar(int64(len(fileSyncReport.Playlists)+1),
			mpb.PrependDecorators(
				// simple name decorator
				decor.Name("Sync playlists"),
				// decor.DSyncWidth bit enables column width synchronization
				decor.Percentage(decor.WCSyncSpace),
			),
		)

		// Delete playlists

		for _, playlistId := range fileSyncReport.DeletedPlaylistIds {
			if synchroAborded {
				break
			}
			func() {
				defer func() {
					select {
					case <-a.interruptRequestChannel:
						synchroAborded = true
						playlistsBar.Abort(false)
					default:
					}
				}()

				if fileSyncLocalPlaylist, ok := a.fileSyncConfig.FileSyncLocalPlaylists[playlistId]; ok {
					fullpath := a.fileSyncMusicFolder + "/playlists/" + fileSyncLocalPlaylist.Filepath
					err := os.Remove(fullpath)
					if err != nil && err != syscall.ENOENT {
						logrus.Warningf("Unable to delete \"%s\": %s\n", fullpath, err.Error())
						playlistSyncErrors++
					} else {
						delete(a.fileSyncConfig.FileSyncLocalPlaylists, playlistId)
						playlistSyncDeletedPlaylists++
						logrus.Debugf("\"%s\" deleted\n", fileSyncLocalPlaylist.Filepath)
					}
				}
			}()
		}

		// Update playlists
		if !synchroAborded {
			playlistsBar.Increment()

			for _, playlist := range fileSyncReport.Playlists {
				if synchroAborded {
					break
				}
				func() {
					defer func() {
						playlistsBar.Increment()
						select {
						case <-a.interruptRequestChannel:
							synchroAborded = true
							playlistsBar.Abort(false)
						default:
						}
					}()

					// Create playlist folder(s)
					newpath := tool.SanitizeFilename(playlist.Name) + ".m3u8"
					fullNewpath := a.fileSyncMusicFolder + "/playlists/" + newpath

					index := strings.LastIndex(fullNewpath, "/")
					if index != -1 {
						err = os.MkdirAll(fullNewpath[:index], 0770)
						if err != nil {
							logrus.Warningf("Unable to create folders for \"%s\": %s\n", newpath, err.Error())
							playlistSyncErrors++
							return
						}
					}

					// Write playlist content
					file, err := os.Create(fullNewpath + ".tmp")
					if err != nil {
						logrus.Warningf("Unable to create file \"%s\": %s\n", fullNewpath+".tmp", err.Error())
						playlistSyncErrors++
						return
					}

					for _, songId := range playlist.SongIds {
						if fileSyncLocalSong, ok := a.fileSyncConfig.FileSyncLocalSongs[songId]; ok {
							fmt.Fprintln(file, "../songs/"+fileSyncLocalSong.Filepath)
						}
					}
					file.Close()

					// Delete old playlist
					if fileSyncLocalPlaylist, ok := a.fileSyncConfig.FileSyncLocalPlaylists[playlist.Id]; ok {
						os.Remove(a.fileSyncMusicFolder + "/playlists/" + fileSyncLocalPlaylist.Filepath)
					}

					// Rename new playlist
					err = os.Rename(fullNewpath+".tmp", fullNewpath)
					if err != nil {
						logrus.Warningf("Unable to rename file \"%s\": %s\n", fullNewpath+".tmp", err.Error())
						playlistSyncErrors++
						return
					}

					// Update playlist sync data
					a.fileSyncConfig.FileSyncLocalPlaylists[playlist.Id] = &FileSyncLocalPlaylist{
						Filepath: newpath,
						UpdateTs: playlist.UpdateTs,
					}
					playlistSyncUpdatedPlaylists++
				}()
			}

		}
	}

	//	progressContainer.Wait()

	if songSyncErrors == 0 && playlistSyncErrors == 0 && !synchroAborded {
		a.fileSyncConfig.LastFileSyncTs = fileSyncReport.SyncTs
	}
	a.saveFileSyncConfig()

	// Report
	if synchroAborded {
		fmt.Println("Synchronization aborded:")
	} else {
		fmt.Println("Synchronization done:")
	}
	fmt.Printf("- Songs: %d updated / %d deleted / %d errors\n", songSyncUpdatedSongs, songSyncDeletedSongs, songSyncErrors)
	fmt.Printf("- FavoritePlaylists: %d updated / %d deleted / %d errors\n", playlistSyncUpdatedPlaylists, playlistSyncDeletedPlaylists, playlistSyncErrors)

	// Cleaning
	select {
	case <-a.interruptRequestChannel:
	default:
	}
}

func deleteVoidParentFolder(filename string) {
	ind := strings.LastIndex(filename, "/")
	if ind != -1 {
		parentFolder := filename[:ind]
		os.Remove(parentFolder)
	}
}

func (a *App) saveFileSyncConfig() {
	logrus.Debugf("Save: %s", a.getCompleteFileSyncFilename())
	rawConfig, err := json.MarshalIndent(a.fileSyncConfig, "", "\t")
	if err != nil {
		logrus.Fatalf("Unable to serialize config file: %v\n", err)
	}
	err = ioutil.WriteFile(a.getCompleteFileSyncFilename(), rawConfig, 0660)
	if err != nil {
		logrus.Fatalf("Unable to save config file: %v\n", err)
	}
}

func (a *App) getCompleteFileSyncFilename() string {
	return a.fileSyncMusicFolder + "/" + FileSyncFilename
}

func (a *App) Init() {
	// Check music folder
	_, err := os.Stat(a.fileSyncMusicFolder)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Creation of music folder: %s\n", a.fileSyncMusicFolder)
			err = os.Mkdir(a.fileSyncMusicFolder, 0770)
			if err != nil {
				logrus.Fatalf("Unable to create music folder: %v\n", err)
			}

		} else {
			logrus.Fatalf("Unable to access music folder: %s", a.fileSyncMusicFolder)
		}
	}

	// Create fileSync file
	a.saveFileSyncConfig()
	fmt.Println("Music folder initialized")

}
