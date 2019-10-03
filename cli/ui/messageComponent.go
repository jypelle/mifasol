package ui

import (
	"github.com/rivo/tview"
)

type MessageComponent struct {
	*tview.TextView
	uiApp *UIApp
}

func NewMessageComponent(uiApp *UIApp) *MessageComponent {

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
