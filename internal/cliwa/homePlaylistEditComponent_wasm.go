package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restApiV1"
)

type HomePlaylistEditComponent struct {
	app          *App
	playlistId   restApiV1.PlaylistId
	playlistMeta *restApiV1.PlaylistMeta
	closed       bool
}

func NewHomePlaylistEditComponent(app *App, playlistId restApiV1.PlaylistId, playlistMeta *restApiV1.PlaylistMeta) *HomePlaylistEditComponent {
	c := &HomePlaylistEditComponent{
		app:          app,
		playlistId:   playlistId,
		playlistMeta: playlistMeta.Copy(),
	}

	return c
}

func (c *HomePlaylistEditComponent) Show() {
	div := jst.Id("homeMainModal")
	div.Set("innerHTML", c.app.RenderTemplate(
		c.playlistMeta, "home/playlistEdit/index"),
	)

	form := jst.Id("playlistEditForm")
	form.Call("addEventListener", "submit", c.app.AddEventFuncPreventDefault(c.saveAction))
	cancelButton := jst.Id("playlistEditCancelButton")
	cancelButton.Call("addEventListener", "click", c.app.AddEventFunc(c.cancelAction))

}

func (c *HomePlaylistEditComponent) saveAction() {
	if c.closed {
		return
	}

	c.app.ShowLoader("Updating playlist")

	playlistName := jst.Id("playlistEditPlaylistName")
	c.playlistMeta.Name = playlistName.Get("value").String()

	if c.playlistId != "" {
		_, cliErr := c.app.restClient.UpdatePlaylist(c.playlistId, c.playlistMeta)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to update the playlist", cliErr)
		}
	} else {
		_, cliErr := c.app.restClient.CreatePlaylist(c.playlistMeta)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to create the playlist", cliErr)
		}
	}
	c.close()
	c.app.HomeComponent.Reload()
	c.app.HideLoader()
}

func (c *HomePlaylistEditComponent) cancelAction() {
	if c.closed {
		return
	}
	c.close()
}

func (c *HomePlaylistEditComponent) close() {
	c.closed = true
	c.app.HomeComponent.CloseModal()
}
