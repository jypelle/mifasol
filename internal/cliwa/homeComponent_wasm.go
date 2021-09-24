package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"strconv"
)

type HomeComponent struct {
	app *App

	MessageComponent *MessageComponent
	LibraryComponent *LibraryComponent
	CurrentComponent *CurrentComponent
	PlayerComponent  *PlayerComponent
}

func NewHomeComponent(app *App) *HomeComponent {
	c := &HomeComponent{
		app: app,
	}

	c.MessageComponent = NewMessageComponent(c.app)
	c.LibraryComponent = NewLibraryComponent(c.app)
	c.CurrentComponent = NewCurrentComponent(c.app)
	c.PlayerComponent = NewPlayerComponent(c.app)

	return c
}

func (c *HomeComponent) Show() {
	body := jst.Document.Get("body")
	body.Set("innerHTML", c.app.RenderTemplate(nil, "home.html"))

	// Set buttons
	logOutButton := jst.Document.Call("getElementById", "logOutButton")
	logOutButton.Call("addEventListener", "click", c.app.AddEventFunc(c.logOutAction))
	refreshButton := jst.Document.Call("getElementById", "refreshButton")
	refreshButton.Call("addEventListener", "click", c.app.AddEventFunc(c.refreshAction))

	c.LibraryComponent.Show()
	c.CurrentComponent.Show()
	c.PlayerComponent.Show()

	c.Reload()

}

func (c *HomeComponent) logOutAction() {
	jst.LocalStorage.Set("mifasolUsername", "")
	jst.LocalStorage.Set("mifasolPassword", "")
	c.app.StartComponent.Show()
}

func (c *HomeComponent) refreshAction() {
	c.Reload()
}

func (c *HomeComponent) Reload() {
	if c.app.localDb == nil {
		return
	}
	c.MessageComponent.Message("Syncing...")
	// Refresh In memory Db
	err := c.app.localDb.Refresh()
	if err != nil {
		c.MessageComponent.Message("Unable to load data from mifasolsrv")
		return
	}

	c.LibraryComponent.RefreshView()
	c.CurrentComponent.RefreshView()

	c.MessageComponent.Message(strconv.Itoa(len(c.app.localDb.Songs)) + " songs, " + strconv.Itoa(len(c.app.localDb.Artists)) + " artists, " + strconv.Itoa(len(c.app.localDb.Albums)) + " albums, " + strconv.Itoa(len(c.app.localDb.Playlists)) + " playlists ready to be played for " + strconv.Itoa(len(c.app.localDb.Users)) + " users.")
}
