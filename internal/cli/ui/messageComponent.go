package ui

import (
	"gitlab.com/tslocum/cview"
)

type MessageComponent struct {
	*cview.TextView
	uiApp *App
}

func NewMessageComponent(uiApp *App) *MessageComponent {

	c := &MessageComponent{
		uiApp: uiApp,
	}

	c.TextView = cview.NewTextView()
	c.TextView.SetDynamicColors(true)
	c.TextView.SetRegions(true)
	c.TextView.SetWrap(false)

	return c
}

func (c *MessageComponent) SetMessage(message string) {
	c.TextView.SetText("[darkcyan]" + message + "[white]")
}

func (c *MessageComponent) SetWarningMessage(message string) {
	c.TextView.SetText("[red]" + message + "[white]")
}
