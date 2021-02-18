package ui

import "github.com/jypelle/mifasol/internal/cli/ui/color"

func (c *PlayerComponent) Enable() {
	c.titleBox.SetBackgroundColor(color.ColorTitleBackground)
	c.progressBox.SetBackgroundColor(color.ColorEnabled)
}

func (c *PlayerComponent) Disable() {
	c.titleBox.SetBackgroundColor(color.ColorTitleUnfocusedBackground)
	c.progressBox.SetBackgroundColor(color.ColorDisabled)
}
