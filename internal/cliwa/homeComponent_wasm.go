package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"strconv"
)

type HomeComponent struct {
	app *App

	HeaderButtonsComponent *HomeHeaderButtonsComponent
	MessageComponent       *HomeMessageComponent
	LibraryComponent       *LibraryComponent
	CurrentComponent       *HomeCurrentComponent
	PlayerComponent        *HomePlayerComponent
}

func NewHomeComponent(app *App) *HomeComponent {
	c := &HomeComponent{
		app: app,
	}

	c.HeaderButtonsComponent = NewHomeHeaderButtonsComponent(c.app)
	c.MessageComponent = NewHomeMessageComponent(c.app)
	c.LibraryComponent = NewHomeLibraryComponent(c.app)
	c.CurrentComponent = NewHomeCurrentComponent(c.app)
	c.PlayerComponent = NewHomePlayerComponent(c.app)

	return c
}

func (c *HomeComponent) Show() {
	mainComponent := jst.Document.Call("getElementById", "mainComponent")
	mainComponent.Set("innerHTML", c.app.RenderTemplate(nil, "home/index"))

	c.HeaderButtonsComponent.Show()
	c.LibraryComponent.Show()
	c.CurrentComponent.Show()
	c.PlayerComponent.Show()

	c.Reload()

}

func (c *HomeComponent) uploadSongsAction() {
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

	c.HeaderButtonsComponent.RefreshView()
	c.LibraryComponent.RefreshView()
	c.CurrentComponent.RefreshView()

	c.MessageComponent.Message(strconv.Itoa(len(c.app.localDb.Songs)) + " songs, " + strconv.Itoa(len(c.app.localDb.Artists)) + " artists, " + strconv.Itoa(len(c.app.localDb.Albums)) + " albums, " + strconv.Itoa(len(c.app.localDb.Playlists)) + " playlists ready to be played for " + strconv.Itoa(len(c.app.localDb.Users)) + " users.")
}

func (c *HomeComponent) CloseModal() {
	homeMainMaster := jst.Document.Call("getElementById", "homeMainMaster")
	homeMainMaster.Get("style").Set("display", "flex")
	homeMainModal := jst.Document.Call("getElementById", "homeMainModal")
	homeMainModal.Set("innerHTML", "")
	homeMainModal.Get("style").Set("display", "none")
	homeHeaderButtonsComponent := jst.Document.Call("getElementById", "homeHeaderButtonsComponent")
	homeHeaderButtonsComponent.Get("style").Set("display", "flex")
}

func (c *HomeComponent) OpenModal() {
	homeMainMaster := jst.Document.Call("getElementById", "homeMainMaster")
	homeMainMaster.Get("style").Set("display", "none")
	homeMainModal := jst.Document.Call("getElementById", "homeMainModal")
	homeMainModal.Get("style").Set("display", "flex")
	homeHeaderButtonsComponent := jst.Document.Call("getElementById", "homeHeaderButtonsComponent")
	homeHeaderButtonsComponent.Get("style").Set("display", "none")
}
