package ui

import (
	"github.com/rivo/tview"
)

type MessageComponent struct {
	*tview.TextView
	uiApp *App
}

func NewMessageComponent(uiApp *App) *MessageComponent {

	c := &MessageComponent{
		uiApp: uiApp,
	}

	c.TextView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	return c
}

func (c *MessageComponent) SetMessage(message string) {
	c.TextView.SetText("[darkcyan]" + message + "[white]")
}

func (c *MessageComponent) SetWarningMessage(message string) {
	c.TextView.SetText("[red]" + message + "[white]")
}
