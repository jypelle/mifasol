package ui

import (
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/rivo/tview"
)

type PlayerComponent struct {
	*tview.Flex
	titleBox    *tview.TextView
	volumeBox   *tview.TextView
	progressBox *tview.Box
	uiApp       *UIApp

	playingSong *restApiV1.Song
}

func NewPlayerComponent(uiApp *UIApp, volume int) *PlayerComponent {
	c := &PlayerComponent{
		uiApp: uiApp,
	}

	c.titleBox = tview.NewTextView()
	c.volumeBox = tview.NewTextView()
	c.progressBox = tview.NewBox()

	c.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(c.progressBox, 1, 0, false)

	return c

}

func (c *PlayerComponent) Focus(delegate func(tview.Primitive)) {
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

func (c *PlayerComponent) Play(songId restApiV1.SongId) {
}
