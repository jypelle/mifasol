package imp

import (
	"fmt"
	"github.com/jypelle/mifasol/internal/cli/config"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/sirupsen/logrus"
	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

type App struct {
	config.ClientConfig
	restClient *restClientV1.RestClient

	importDir                       string
	importOneFolderPerAlbumDisabled bool

	doneChannel             chan bool
	interruptChannel        chan os.Signal
	interruptRequestChannel chan bool
}

func NewApp(clientConfig config.ClientConfig, restClient *restClientV1.RestClient, importDir string, importOneFolderPerAlbumDisabled bool) *App {
	app := &App{
		ClientConfig:                    clientConfig,
		restClient:                      restClient,
		importDir:                       importDir,
		importOneFolderPerAlbumDisabled: importOneFolderPerAlbumDisabled,
		doneChannel:                     make(chan bool),
		interruptChannel:                make(chan os.Signal),
		interruptRequestChannel:         make(chan bool),
	}

	return app
}

func (a *App) Start() {
	signal.Notify(a.interruptChannel, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)

	go a.start()

	select {
	case <-a.doneChannel:
		return
	case <-a.interruptChannel:
		fmt.Println("Interruption requested: finishing current file upload before exiting\n")
		a.interruptRequestChannel <- true
		<-a.doneChannel
	}

}

func (a *App) start() {
	defer func() { a.doneChannel <- true }()

	impAborded := false

	fmt.Printf("Scanning folder \"%s\"\n", a.importDir)

	var filesNameToImport []string
	var filesSongFormatToImport []restApiV1.SongFormat
	var filesSize []int64

	// Identify every song files to import
	err := filepath.Walk(a.importDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			lowerCasePath := strings.ToLower(path)

			if !info.IsDir() && info.Size() > 20000 {
				songFormat := restApiV1.SongFormatUnknown

				switch {
				case strings.HasSuffix(lowerCasePath, ".flac"):
					logrus.Debugf("Detect flac file: %s", path)
					songFormat = restApiV1.SongFormatFlac

				case strings.HasSuffix(lowerCasePath, ".mp3"):
					logrus.Debugf("Detect mp3 file: %s", path)
					songFormat = restApiV1.SongFormatMp3
					/*
						case strings.HasSuffix(lowerCasePath, ".ogg") || strings.HasSuffix(lowerCasePath, ".oga"):
							logrus.Debugf("Detect ogg file: %s", path)
							songFormat = model.SongFormatOgg
					*/
				}

				if songFormat != restApiV1.SongFormatUnknown {
					filesNameToImport = append(filesNameToImport, path)
					filesSongFormatToImport = append(filesSongFormatToImport, songFormat)
					filesSize = append(filesSize, info.Size())
				}
			}
			return nil
		})

	if err != nil {
		logrus.Fatalf("Unable to parse importing folder: %v", err)
	}

	// Try to import every song files previously identified
	importedSongs := 0

	if len(filesNameToImport) == 0 {
		fmt.Println("No files to import")
	} else {
		fmt.Printf("Trying to import %d songs\n", len(filesNameToImport))

		progressContainer := mpb.New()
		bar := progressContainer.AddBar(int64(len(filesNameToImport)),
			mpb.PrependDecorators(
				// simple name decorator
				decor.Name("Import songs"),
				// decor.DSyncWidth bit enables column width synchronization
				decor.Percentage(decor.WCSyncSpace),
			),
		)

		var lastFolder string
		var lastAlbumId restApiV1.AlbumId = restApiV1.UnknownAlbumId

		for key, fileName := range filesNameToImport {
			if impAborded {
				break
			}
			func() {
				defer func() {
					select {
					case <-a.interruptRequestChannel:
						impAborded = true
						bar.Abort(false)
					default:
					}
				}()

				indexLastSeparator := strings.LastIndex(fileName, string(filepath.Separator))

				// Reset last album id on new folder
				if lastFolder != fileName[:indexLastSeparator] {
					lastAlbumId = restApiV1.UnknownAlbumId
				}

				songFormat := filesSongFormatToImport[key]

				var err error
				var reader *os.File
				reader, err = os.Open(fileName)
				if err == nil {
					defer reader.Close()

					truncatedFilePath := tool.CharacterTruncate(fileName[indexLastSeparator+1:], 23)

					songBar := progressContainer.AddBar(filesSize[key],
						mpb.PrependDecorators(
							decor.Name("Uploading   "),
							decor.Percentage(decor.WCSyncSpace),
						),
						mpb.AppendDecorators(
							decor.CountersKibiByte("%6.1f / %6.1f"),
							decor.Name(" "+truncatedFilePath),
						),
						mpb.BarRemoveOnComplete(),
					)

					proxyReader := songBar.ProxyReader(reader)

					var apiErr restClientV1.ClientError
					var song *restApiV1.Song

					if a.importOneFolderPerAlbumDisabled {
						song, apiErr = a.restClient.CreateSongContent(songFormat, proxyReader)
					} else {
						song, apiErr = a.restClient.CreateSongContentForAlbum(songFormat, proxyReader, lastAlbumId)
					}
					if apiErr == nil {
						importedSongs++
						lastAlbumId = song.AlbumId
					} else {
						songBar.Abort(true)
						logrus.Warnf("Unable to import file %s: %v", fileName, apiErr)
					}
				} else {
					logrus.Warnf("Unable to upload file %s: %v", fileName, err)
				}

				lastFolder = fileName[:indexLastSeparator]
				bar.Increment()

			}()
		}
		progressContainer.Wait()

	}

	// Report
	if impAborded {
		fmt.Print("Import aborded: ")
	} else {
		fmt.Print("Import done: ")
	}
	fmt.Printf("%d songs imported\n", importedSongs)

	// Cleaning
	select {
	case <-a.interruptRequestChannel:
	default:
	}
}
