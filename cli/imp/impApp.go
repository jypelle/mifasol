package imp

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
	"lyra/cli/config"
	"lyra/restApiV1"
	"lyra/restClientV1"
	"lyra/tool"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

type ImpApp struct {
	config.ClientConfig
	restClient *restClientV1.RestClient

	importDir string

	doneChannel             chan bool
	interruptChannel        chan os.Signal
	interruptRequestChannel chan bool
}

func NewImpApp(clientConfig config.ClientConfig, restClient *restClientV1.RestClient, importDir string) *ImpApp {
	impApp := &ImpApp{
		ClientConfig:            clientConfig,
		restClient:              restClient,
		importDir:               importDir,
		doneChannel:             make(chan bool),
		interruptChannel:        make(chan os.Signal),
		interruptRequestChannel: make(chan bool),
	}

	return impApp
}

func (a *ImpApp) Start() {
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

func (a *ImpApp) start() {
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

			songFormat := filesSongFormatToImport[key]

			var err error
			var reader *os.File
			reader, err = os.Open(fileName)
			if err == nil {
				defer reader.Close()

				indexLastSeparator := strings.LastIndex(fileName, "/")
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

				_, apiErr := a.restClient.CreateSongContent(songFormat, proxyReader)
				if apiErr == nil {
					importedSongs++
				}
			} else {
				logrus.Warnf("Unable to import file %s: %v", fileName, err)
			}
			bar.Increment()

		}()
	}
	progressContainer.Wait()

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
