package cliwa

import (
	"fmt"
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restApiV1"
	"html"
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
	player := jst.Document.Call("getElementById", "playerAudio")
	player.Call("addEventListener", "ended", c.app.AddEventFunc(c.app.HomeComponent.CurrentComponent.PlayNextSongAction))

	playerPlayButton := jst.Document.Call("getElementById", "playerPlayButton")
	playerPlayButton.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		playerAudio := jst.Document.Call("getElementById", "playerAudio")
		if playerAudio.Get("paused").Bool() {
			c.ResumeSongAction()
		} else {
			c.PauseSongAction()
		}
	}))
}

func (c *PlayerComponent) PlaySongAction(songId restApiV1.SongId) {
	token, cliErr := c.app.restClient.GetToken()

	if cliErr != nil {
		return
	}

	playerPlayButton := jst.Document.Call("getElementById", "playerPlayButton")
	playerPlayButton.Set("innerHTML", `<i class="fas fa-pause"></i>`)

	player := jst.Document.Call("getElementById", "playerAudio")
	player.Set("src", "/api/v1/songContents/"+string(songId)+"?bearer="+token.AccessToken)
	player.Call("play")

	c.app.HomeComponent.MessageComponent.Message(fmt.Sprintf(`Playing <span class="songLink">%s</span> ...`, html.EscapeString(c.app.localDb.Songs[songId].Name)))

	return
}

func (c *PlayerComponent) PauseSongAction() {
	playerPlayButton := jst.Document.Call("getElementById", "playerPlayButton")
	playerPlayButton.Set("innerHTML", `<i class="fas fa-play"></i>`)
	player := jst.Document.Call("getElementById", "playerAudio")
	player.Call("pause")
}

func (c *PlayerComponent) ResumeSongAction() {
	playerPlayButton := jst.Document.Call("getElementById", "playerPlayButton")
	playerPlayButton.Set("innerHTML", `<i class="fas fa-pause"></i>`)
	player := jst.Document.Call("getElementById", "playerAudio")
	player.Call("play")
}
