package ui

import "github.com/gdamore/tcell"

func (c *PlayerComponent) Enable() {
	c.titleBox.SetBackgroundColor(tcell.NewHexColor(0xe0e0e0))
	c.progressBox.SetBackgroundColor(ColorEnabled)
}

func (c *PlayerComponent) Disable() {
	c.titleBox.SetBackgroundColor(tcell.NewHexColor(0xc0c0c0))
	c.progressBox.SetBackgroundColor(ColorDisabled)
}
