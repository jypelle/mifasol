package ui

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
	"github.com/gdamore/tcell"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/rivo/tview"
	"time"
)

type PlayerComponent struct {
	*tview.Flex
	titleBox    *tview.TextView
	volumeBox   *tview.TextView
	progressBox *tview.Box
	uiApp       *UIApp

	volume          int
	musicStreamer   beep.StreamSeekCloser
	musicFormat     beep.Format
	controlStreamer *beep.Ctrl
	volumeStreamer  *effects.Volume

	playingSong *restApiV1.Song
}

func NewPlayerComponent(uiApp *UIApp, volume int) *PlayerComponent {

	// Init soundcard
	var sampleRate beep.SampleRate = 44100
	speaker.Init(sampleRate, sampleRate.N(time.Duration(uiApp.BufferLength)*time.Millisecond))

	c := &PlayerComponent{
		uiApp: uiApp,
	}

	c.titleBox = tview.NewTextView()
	c.titleBox.SetDynamicColors(true)
	c.titleBox.SetText("[" + ColorTitleStr + "]  ...")

	c.volumeBox = tview.NewTextView()
	c.volumeBox.SetDynamicColors(true)
	c.SetVolume(volume)
	c.volumeBox.SetBackgroundColor(ColorTitleBackground)
	c.volumeBox.SetTextAlign(tview.AlignRight)

	c.progressBox = tview.NewBox()

	c.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexColumn).
			AddItem(c.titleBox, 0, 1, false).
			AddItem(c.volumeBox, 7, 0, false),
			1, 0, false).
		AddItem(c.progressBox, 1, 0, false)

	c.progressBox.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {

		case event.Key() == tcell.KeyRight:
			speaker.Lock()
			if c.musicStreamer != nil {
				newPosition := c.musicStreamer.Position() + c.musicFormat.SampleRate.N(5*time.Second)
				if newPosition < c.musicStreamer.Len() {
					c.musicStreamer.Seek(newPosition)
				}
			}
			speaker.Unlock()
			return nil
		case event.Key() == tcell.KeyLeft:
			speaker.Lock()
			if c.musicStreamer != nil {
				newPosition := c.musicStreamer.Position() - c.musicFormat.SampleRate.N(5*time.Second)
				if newPosition < 0 {
					newPosition = 0
				}
				c.musicStreamer.Seek(newPosition)
				//				c.musicStreamer.Seek(0)
			}
			speaker.Unlock()
			return nil
		}

		return event
	})

	return c

}

func (c *PlayerComponent) Focus(delegate func(tview.Primitive)) {
	delegate(c.progressBox)
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
		c.volumeBox.SetText("[" + ColorTitleStr + "]🔇" + fmt.Sprintf("%3d", c.volume) + "%")
	} else {
		c.volumeBox.SetText("[" + ColorTitleStr + "]🔉" + fmt.Sprintf("%3d", c.volume) + "%")
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
			c.titleBox.SetText("[" + ColorTitleStr + "]Paused: " + c.getMainTextSong(c.playingSong))
		} else {
			c.titleBox.SetText("[" + ColorTitleStr + "]Playing: " + c.getMainTextSong(c.playingSong))
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

func (c *PlayerComponent) Play(songId restApiV1.SongId) {
	song, ok := c.uiApp.localDb.Songs[songId]
	if !ok {
		c.uiApp.WarningMessage("Unknown song id: " + string(songId))
		return
	}

	c.playingSong = song

	c.uiApp.Message("Start playing: " + c.getMainTextSong(c.playingSong))
	c.uiApp.tviewApp.ForceDraw()

	reader, _, cliErr := c.uiApp.restClient.ReadSongContent(song.Id)
	if cliErr != nil {
		c.uiApp.ClientErrorMessage("Unable to retrieve content for: "+c.playingSong.Name, cliErr)
		return
	}

	var err error
	c.musicStreamer, c.musicFormat, err = song.Format.Decode()(reader)
	if err != nil {
		c.uiApp.WarningMessage("Unable to read content for: " + c.playingSong.Name)
		return
	}

	speaker.Clear()

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
					c.uiApp.tviewApp.QueueUpdateDraw(func() {
						c.titleBox.SetText("[" + ColorTitleStr + "]Stopped: " + c.getMainTextSong(c.playingSong))
						nextSongId := c.uiApp.currentComponent.GetNextSong()
						if nextSongId != nil {
							c.Play(*nextSongId)
						}
					})
				},
			),
		),
	)

	c.titleBox.SetText("[" + ColorTitleStr + "]Playing: " + c.getMainTextSong(c.playingSong))
	c.uiApp.tviewApp.ForceDraw()

}

func (c *PlayerComponent) getMainTextSong(song *restApiV1.Song) string {
	songName := tview.Escape(song.Name)

	albumName := ""
	if song.AlbumId != restApiV1.UnknownAlbumId {
		albumName = " / " + tview.Escape(c.uiApp.localDb.Albums[song.AlbumId].Name)
	}
	artistsName := ""
	if len(song.ArtistIds) > 0 {
		for _, artistId := range song.ArtistIds {
			artistsName += " / " + tview.Escape(c.uiApp.localDb.Artists[artistId].Name)
		}
	}
	return songName + albumName + artistsName
}
