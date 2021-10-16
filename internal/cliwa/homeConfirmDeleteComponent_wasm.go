package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restApiV1"
)

type HomeConfirmDeleteComponent struct {
	app    *App
	id     interface{}
	name   string
	closed bool
}

func NewHomeConfirmDeleteComponent(
	app *App,
	id interface{},
) *HomeConfirmDeleteComponent {
	c := &HomeConfirmDeleteComponent{
		app: app,
		id:  id,
	}

	switch v := id.(type) {
	case restApiV1.ArtistId:
		c.name = app.localDb.Artists[v].Name
	case restApiV1.AlbumId:
		c.name = app.localDb.Albums[v].Name
	case restApiV1.SongId:
		c.name = app.localDb.Songs[v].Name
	case restApiV1.PlaylistId:
		c.name = app.localDb.Playlists[v].Name
	case restApiV1.UserId:
		c.name = app.localDb.Users[v].Name
	default:
		return nil
	}

	return c
}

func (c *HomeConfirmDeleteComponent) Show() {
	div := jst.Document.Call("getElementById", "homeMainModal")
	div.Set("innerHTML", c.app.RenderTemplate(
		c.name, "home/confirmDelete/index"),
	)

	form := jst.Document.Call("getElementById", "confirmDeleteForm")
	form.Call("addEventListener", "submit", c.app.AddEventFuncPreventDefault(c.deleteAction))
	cancelButton := jst.Document.Call("getElementById", "confirmDeleteCancelButton")
	cancelButton.Call("addEventListener", "click", c.app.AddEventFunc(c.cancelAction))

}

func (c *HomeConfirmDeleteComponent) deleteAction() {
	if c.closed {
		return
	}

	switch v := c.id.(type) {
	case restApiV1.ArtistId:
		c.app.ShowLoader("Deleting the artist")
		_, cliErr := c.app.restClient.DeleteArtist(v)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to delete the artist", cliErr)
		}
	case restApiV1.AlbumId:
		c.app.ShowLoader("Deleting the album")
		_, cliErr := c.app.restClient.DeleteAlbum(v)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to delete the album", cliErr)
		}
	case restApiV1.SongId:
		c.app.ShowLoader("Deleting the song")
		_, cliErr := c.app.restClient.DeleteSong(v)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to delete the song", cliErr)
		}
	case restApiV1.PlaylistId:
		c.app.ShowLoader("Deleting the playlist")
		_, cliErr := c.app.restClient.DeletePlaylist(v)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to delete the playlist", cliErr)
		}
	case restApiV1.UserId:
		c.app.ShowLoader("Deleting the user")
		_, cliErr := c.app.restClient.DeleteUser(v)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to delete the user", cliErr)
		}
	}

	c.close()
	c.app.HomeComponent.Reload()
	c.app.HideLoader()
}

func (c *HomeConfirmDeleteComponent) cancelAction() {
	if c.closed {
		return
	}
	c.close()
}

func (c *HomeConfirmDeleteComponent) close() {
	c.closed = true
	c.app.HomeComponent.CloseModal()
}
