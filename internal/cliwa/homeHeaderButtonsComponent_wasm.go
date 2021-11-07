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
	div := jst.Id("homeHeaderButtonsComponent")
	div.Set("innerHTML", c.app.RenderTemplate(
		nil, "home/headerButtons/index"),
	)

	// Set buttons
	uploadSongsButton := jst.Id("uploadSongsButton")
	uploadSongsButton.Call("addEventListener", "click", c.app.AddEventFunc(c.app.HomeComponent.uploadSongsAction))
	logOutButton := jst.Id("logOutButton")
	logOutButton.Call("addEventListener", "click", c.app.AddEventFunc(c.app.HomeComponent.logOutAction))
	refreshButton := jst.Id("refreshButton")
	refreshButton.Call("addEventListener", "click", c.app.AddEventFunc(c.app.HomeComponent.refreshAction))
}

func (c *HomeHeaderButtonsComponent) RefreshView() {
	uploadSongsButton := jst.Id("uploadSongsButton")
	if c.app.IsConnectedUserAdmin() {
		uploadSongsButton.Set("style", "display:block;")
	} else {
		uploadSongsButton.Set("style", "display:none;")
	}
}
