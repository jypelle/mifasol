package ui

import (
	"code.rocketnine.space/tslocum/cview"
	"github.com/jypelle/mifasol/internal/cli/ui/color"
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
	c.TextView.SetBackgroundColor(color.ColorEnabled)

	return c
}

func (c *MessageComponent) SetMessage(message string) {
	c.TextView.SetText("[darkcyan]" + message + "[" + color.ColorWhiteStr + "]")
}

func (c *MessageComponent) SetWarningMessage(message string) {
	c.TextView.SetText("[red]" + message + "[" + color.ColorWhiteStr + "]")
}
