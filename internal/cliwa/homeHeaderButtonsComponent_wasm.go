package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
)

type HomeHeaderButtonsComponent struct {
	app *App
}

func NewHomeHeaderButtonsComponent(app *App) *HomeHeaderButtonsComponent {
	c := &HomeHeaderButtonsComponent{
		app: app,
	}

	return c
}

func (c *HomeHeaderButtonsComponent) Show() {
	div := jst.Document.Call("getElementById", "homeHeaderButtonsComponent")
	div.Set("innerHTML", c.app.RenderTemplate(
		nil, "homeHeaderButtons.html"),
	)

	// Set buttons
	uploadSongsButton := jst.Document.Call("getElementById", "uploadSongsButton")
	uploadSongsButton.Call("addEventListener", "click", c.app.AddEventFunc(c.app.HomeComponent.uploadSongsAction))
	logOutButton := jst.Document.Call("getElementById", "logOutButton")
	logOutButton.Call("addEventListener", "click", c.app.AddEventFunc(c.app.HomeComponent.logOutAction))
	refreshButton := jst.Document.Call("getElementById", "refreshButton")
	refreshButton.Call("addEventListener", "click", c.app.AddEventFunc(c.app.HomeComponent.refreshAction))
}

func (c *HomeHeaderButtonsComponent) RefreshView() {
	uploadSongsButton := jst.Document.Call("getElementById", "uploadSongsButton")
	if c.app.IsConnectedUserAdmin() {
		uploadSongsButton.Set("style", "display:block;")
	} else {
		uploadSongsButton.Set("style", "display:none;")
	}
}
