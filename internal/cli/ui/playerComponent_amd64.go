package ui

import (
	"code.rocketnine.space/tslocum/cview"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/jypelle/mifasol/internal/cli/ui/color"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"io"
	"strconv"
	"time"
)

type PlayerComponent struct {
	*cview.Flex
	titleBox    *cview.TextView
	volumeBox   *cview.TextView
	progressBox *cview.TextView
	uiApp       *App

	volume          int
	musicStreamer   beep.StreamSeekCloser
	musicFormat     beep.Format
	controlStreamer *beep.Ctrl
	volumeStreamer  *effects.Volume

	refreshTicker *time.Ticker

	playingSong *restApiV1.Song
}

func NewPlayerComponent(uiApp *App, volume int) *PlayerComponent {

	// Init soundcard
	var sampleRate beep.SampleRate = 44100
	speaker.Init(sampleRate, sampleRate.N(time.Duration(uiApp.BufferLength)*time.Millisecond))

	c := &PlayerComponent{
		uiApp: uiApp,
	}

	c.titleBox = cview.NewTextView()
	c.titleBox.SetDynamicColors(true)

	c.titleBox.SetText("[" + color.ColorTitleStr + "]  ...")
	c.titleBox.SetBackgroundColor(color.ColorTitleUnfocusedBackground)

	c.progressBox = cview.NewTextView()
	c.progressBox.SetDynamicColors(true)
	c.progressBox.SetText("[" + color.ColorTitleStr + "]00:00 / 00:00")
	c.progressBox.SetBackgroundColor(color.ColorTitleBackground)
	c.progressBox.SetTextAlign(cview.AlignRight)

	c.volumeBox = cview.NewTextView()
	c.volumeBox.SetDynamicColors(true)
	c.SetVolume(volume)
	c.volumeBox.SetBackgroundColor(color.ColorTitleBackground)
	c.volumeBox.SetTextAlign(cview.AlignRight)

	c.Flex = cview.NewFlex()
	c.Flex.SetDirection(cview.FlexColumn)
	c.Flex.AddItem(c.titleBox, 0, 1, false)
	c.Flex.AddItem(c.progressBox, 14, 0, false)
	c.Flex.AddItem(c.volumeBox, 7, 0, false)

	c.refreshTicker = time.NewTicker(time.Second)

	go func() {
		for {
			select {
			case <-c.refreshTicker.C:
				c.refreshProgress()
			}
		}
	}()

	return c

}

func (c *PlayerComponent) SetVolume(volume int) {

	if volume > 120 {
		c.volume = 120
	} else if volume < 0 {
		c.volume = 0
	} else {
		c.volume = volume
	}

	speaker.Lock()
	if c.volumeStreamer != nil {
		c.volumeStreamer.Volume = float64(c.volume-100) / 16
		c.volumeStreamer.Silent = c.volume == 0
	}
	speaker.Unlock()

	if c.volume == 0 {
		c.volumeBox.SetText("[" + color.ColorTitleStr + "]🔇" + fmt.Sprintf("%3d", c.volume) + "%")
	} else {
		c.volumeBox.SetText("[" + color.ColorTitleStr + "]🔉" + fmt.Sprintf("%3d", c.volume) + "%")
	}
}

func (c *PlayerComponent) PauseResume() {
	if c.playingSong != nil {
		speaker.Lock()
		if c.controlStreamer != nil {
			c.controlStreamer.Paused = !c.controlStreamer.Paused
		}
		speaker.Unlock()
		if c.controlStreamer.Paused {
			c.titleBox.SetText("[" + color.ColorTitleStr + "]Paused: " + c.getCompleteMainTextSong(c.playingSong))
		} else {
			c.titleBox.SetText("[" + color.ColorTitleStr + "]Playing: " + c.getCompleteMainTextSong(c.playingSong))
		}
	}
}

func (c *PlayerComponent) GetVolume() int {
	return c.volume
}

func (c *PlayerComponent) VolumeUp() {
	c.SetVolume(c.GetVolume() + 4)
}

func (c *PlayerComponent) VolumeDown() {
	c.SetVolume(c.GetVolume() - 4)
}

func (c *PlayerComponent) GoBackward() {
	speaker.Lock()

	if c.musicStreamer != nil {
		newPosition := c.musicStreamer.Position() - c.musicFormat.SampleRate.N(5*time.Second)
		if newPosition < 0 {
			newPosition = 0
		}
		c.musicStreamer.Seek(newPosition)
	}
	speaker.Unlock()
	c.refreshProgress()
}

func (c *PlayerComponent) GoForward() {
	speaker.Lock()
	if c.musicStreamer != nil {
		currentPosition := c.musicStreamer.Position()
		newPosition := currentPosition + c.musicFormat.SampleRate.N(5*time.Second)
		len := c.musicStreamer.Len()
		if newPosition < len {
			c.musicStreamer.Seek(newPosition)
		}
	}
	speaker.Unlock()
	c.refreshProgress()
}

func (c *PlayerComponent) Play(songId restApiV1.SongId) {
	song, ok := c.uiApp.localDb.Songs[songId]
	if !ok {
		c.uiApp.WarningMessage("Unknown song id: " + string(songId))
		return
	}

	c.playingSong = song

	c.uiApp.Message("Start playing: " + c.getMainTextSong(c.playingSong))
	c.uiApp.cviewApp.Draw()

	songReader, songSize, cliErr := c.uiApp.restClient.ReadSongContent(song.Id)
	if cliErr != nil {
		c.uiApp.ClientErrorMessage("Unable to retrieve content for: "+c.playingSong.Name, cliErr)
		return
	}

	speaker.Clear()

	bufferedReader := tool.NewBufferedStreamReader(songReader, int(songSize), 8192)

	var err error
	var decoder func(rc io.ReadCloser) (s beep.StreamSeekCloser, format beep.Format, err error)

	switch song.Format {
	case restApiV1.SongFormatFlac:
		decoder = func(rc io.ReadCloser) (s beep.StreamSeekCloser, format beep.Format, err error) {
			return flac.Decode(rc)
		}
	case restApiV1.SongFormatOgg:
		decoder = vorbis.Decode
	case restApiV1.SongFormatMp3:
		decoder = mp3.Decode
	default:
		c.uiApp.WarningMessage("Unknown format: " + song.Format.String())
		return
	}

	c.musicStreamer, c.musicFormat, err = decoder(bufferedReader)
	if err != nil {
		c.uiApp.WarningMessage("Unable to read content for: " + c.playingSong.Name)
		return
	}

	if c.musicFormat.SampleRate == 44100 {
		c.controlStreamer = &beep.Ctrl{Streamer: c.musicStreamer, Paused: false}
	} else {
		c.controlStreamer = &beep.Ctrl{Streamer: beep.Resample(4, c.musicFormat.SampleRate, 44100, c.musicStreamer), Paused: false}
	}

	c.volumeStreamer = &effects.Volume{
		Streamer: c.controlStreamer,
		Base:     2,
		Volume:   float64(c.volume-100) / 16,
		Silent:   c.volume == 0,
	}

	speaker.Play(
		beep.Seq(
			c.volumeStreamer,
			beep.Callback(
				func() {
					c.uiApp.cviewApp.QueueUpdateDraw(func() {
						c.titleBox.SetText("[" + color.ColorTitleStr + "]Stopped: " + c.getCompleteMainTextSong(c.playingSong))
						c.refreshProgress()
						c.controlStreamer = nil
						c.volumeStreamer = nil
						c.musicStreamer = nil
						nextSongId := c.uiApp.currentComponent.GetNextSong()
						if nextSongId != nil {
							c.Play(*nextSongId)
						}
					})
				},
			),
		),
	)

	c.titleBox.SetText("[" + color.ColorTitleStr + "]Playing: " + c.getCompleteMainTextSong(c.playingSong))
	c.uiApp.cviewApp.Draw()

}

func (c *PlayerComponent) getMainTextSong(song *restApiV1.Song) string {
	songName := cview.Escape(song.Name)

	return songName
}

func (c *PlayerComponent) getCompleteMainTextSong(song *restApiV1.Song) string {
	songName := cview.Escape(song.Name)

	albumName := ""
	if song.AlbumId != restApiV1.UnknownAlbumId {
		albumName = " / " + cview.Escape(c.uiApp.localDb.Albums[song.AlbumId].Name)
	}
	artistsName := ""
	if len(song.ArtistIds) > 0 {
		for _, artistId := range song.ArtistIds {
			artistsName += " / " + cview.Escape(c.uiApp.localDb.Artists[artistId].Name)
		}
	}
	format := cview.Escape(" [" + song.Format.String() + "/" + song.BitDepth.String() + "/" + strconv.Itoa(int(c.musicFormat.SampleRate)) + "hz]")

	return songName + albumName + artistsName + format
}

func (c *PlayerComponent) refreshProgress() {
	speaker.Lock()
	if c.controlStreamer != nil {
		duration := c.musicFormat.SampleRate.D(c.musicStreamer.Position()).Round(time.Second)
		min := duration / time.Minute
		duration -= min * time.Minute
		sec := duration / time.Second

		duration = c.musicFormat.SampleRate.D(c.musicStreamer.Len()).Round(time.Second)
		totalMin := duration / time.Minute
		duration -= totalMin * time.Minute
		totalSec := duration / time.Second

		c.progressBox.SetText("[" + color.ColorTitleStr + "]" + fmt.Sprintf("%02d:%02d / %02d:%02d", min, sec, totalMin, totalSec))
		c.uiApp.cviewApp.Draw()

	}
	speaker.Unlock()
}
