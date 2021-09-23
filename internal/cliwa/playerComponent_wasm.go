package cliwa

import (
	"fmt"
	"github.com/jypelle/mifasol/restApiV1"
)

type PlayerComponent struct {
	app *App
}

func NewPlayerComponent(app *App) *PlayerComponent {
	c := &PlayerComponent{
		app: app,
	}

	return c
}

func (c *PlayerComponent) Show() {
	player := c.app.doc.Call("getElementById", "player")
	player.Call("addEventListener", "ended", c.app.AddEventFunc(c.app.HomeComponent.CurrentComponent.PlayNextSongAction))
}

func (c *PlayerComponent) PlaySongAction(songId restApiV1.SongId) {
	token, cliErr := c.app.restClient.GetToken()

	if cliErr != nil {
		return
	}

	player := c.app.doc.Call("getElementById", "player")
	player.Set("src", "/api/v1/songContents/"+string(songId)+"?bearer="+token.AccessToken)
	player.Call("play")

	c.app.HomeComponent.MessageComponent.Message(fmt.Sprintf("Playing %s ...", c.app.localDb.Songs[songId].Name))
	return
}

func (c *PlayerComponent) PauseSongAction() {
	player := c.app.doc.Call("getElementById", "player")
	player.Call("pause")
}
