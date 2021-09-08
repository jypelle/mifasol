package cliwa

import "github.com/jypelle/mifasol/restApiV1"

type CurrentComponent struct {
	app *App

	songIds       []restApiV1.SongId
	srcPlaylistId *restApiV1.PlaylistId
	modified      bool
}

func NewCurrentComponent(app *App) *CurrentComponent {
	c := &CurrentComponent{
		app: app,
	}

	return c
}

func (c *CurrentComponent) showCurrentComponent() {
	listDiv := c.app.doc.Call("getElementById", "currentList")

	listDiv.Set("innerHTML", "...")
}
