package ui

func (c *PlayerComponent) Enable() {
	c.titleBox.SetBackgroundColor(ColorTitleBackground)
	c.progressBox.SetBackgroundColor(ColorEnabled)
}

func (c *PlayerComponent) Disable() {
	c.titleBox.SetBackgroundColor(ColorTitleUnfocusedBackground)
	c.progressBox.SetBackgroundColor(ColorDisabled)
}
