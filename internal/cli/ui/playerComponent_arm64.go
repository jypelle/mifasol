package ui

import (
	"code.rocketnine.space/tslocum/cview"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/jypelle/mifasol/restApiV1"
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
	c := &PlayerComponent{
		uiApp: uiApp,
	}

	c.titleBox = cview.NewTextView()
	c.volumeBox = cview.NewTextView()
	c.progressBox = cview.NewTextView()

	c.Flex = cview.NewFlex()
	c.Flex.SetDirection(cview.FlexRow)
	c.Flex.AddItem(c.progressBox, 1, 0, false)

	return c

}

func (c *PlayerComponent) Focus(delegate func(cview.Primitive)) {
	delegate(c.progressBox)
}

func (c *PlayerComponent) SetVolume(volume int) {
}

func (c *PlayerComponent) PauseResume() {
}

func (c *PlayerComponent) GetVolume() int {
	return 0
}

func (c *PlayerComponent) VolumeUp() {
}

func (c *PlayerComponent) VolumeDown() {
}

func (c *PlayerComponent) GoBackward() {
}

func (c *PlayerComponent) GoForward() {
}

func (c *PlayerComponent) Play(songId restApiV1.SongId) {
}
