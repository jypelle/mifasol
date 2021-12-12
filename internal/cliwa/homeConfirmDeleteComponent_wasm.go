package cliwa

import (
	"fmt"
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restApiV1"
	"html"
	"html/template"
)

type HomeConfirmDeleteComponent struct {
	app    *App
	id     interface{}
	name   template.HTML
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
		if len(app.localDb.ArtistOrderedSongs[v]) == 1 {
			c.name = template.HTML(fmt.Sprintf(
				"Do you want to delete <span class=\"artistLink\">%s</span> and its song ?",
				html.EscapeString(app.localDb.Artists[v].Name),
			))
		} else if len(app.localDb.ArtistOrderedSongs[v]) > 1 {
			c.name = template.HTML(fmt.Sprintf(
				"Do you want to delete <span class=\"artistLink\">%s</span> and its %d songs ?",
				html.EscapeString(app.localDb.Artists[v].Name),
				len(app.localDb.ArtistOrderedSongs[v]),
			))
		} else {
			c.name = template.HTML(fmt.Sprintf(
				"Do you want to delete <span class=\"artistLink\">%s</span> ?", html.EscapeString(app.localDb.Artists[v].Name),
			))
		}
	case restApiV1.AlbumId:
		if len(app.localDb.AlbumOrderedSongs[v]) == 1 {
			c.name = template.HTML(fmt.Sprintf(
				"Do you want to delete <span class=\"albumLink\">%s</span> and its song ?",
				html.EscapeString(app.localDb.Albums[v].Name),
			))
		} else if len(app.localDb.AlbumOrderedSongs[v]) > 1 {
			c.name = template.HTML(fmt.Sprintf(
				"Do you want to delete <span class=\"albumLink\">%s</span> and its %d songs ?",
				html.EscapeString(app.localDb.Albums[v].Name),
				len(app.localDb.AlbumOrderedSongs[v]),
			))
		} else {
			c.name = template.HTML(fmt.Sprintf(
				"Do you want to delete <span class=\"albumLink\">%s</span> ?", html.EscapeString(app.localDb.Albums[v].Name),
			))
		}
	case restApiV1.SongId:
		c.name = template.HTML(fmt.Sprintf(
			"Do you want to delete <span class=\"songLink\">%s</span> ?", html.EscapeString(app.localDb.Songs[v].Name),
		))
	case restApiV1.PlaylistId:
		c.name = template.HTML(fmt.Sprintf(
			"Do you want to delete <span class=\"playlistLink\">%s</span> ?", html.EscapeString(app.localDb.Playlists[v].Name),
		))
	case restApiV1.UserId:
		c.name = template.HTML(fmt.Sprintf(
			"Do you want to delete <span class=\"userLink\">%s</span> ?", html.EscapeString(app.localDb.Users[v].Name),
		))
	default:
		return nil
	}

	return c
}

func (c *HomeConfirmDeleteComponent) Render() {
	div := jst.Id("homeMainModal")
	div.Set("innerHTML", c.app.RenderTemplate(
		c.name, "home/confirmDelete/index"),
	)

	form := jst.Id("confirmDeleteForm")
	form.Call("addEventListener", "submit", c.app.AddEventFuncPreventDefault(c.deleteAction))
	cancelButton := jst.Id("confirmDeleteCancelButton")
	cancelButton.Call("addEventListener", "click", c.app.AddEventFunc(c.cancelAction))

}

func (c *HomeConfirmDeleteComponent) deleteAction() {
	if c.closed {
		return
	}
	c.close()

	defer c.app.HideLoader()

	switch v := c.id.(type) {
	case restApiV1.ArtistId:
		artist := c.app.localDb.Artists[v]
		for _, song := range c.app.localDb.ArtistOrderedSongs[v] {
			c.app.ShowLoader(fmt.Sprintf("Deleting <span class=\"songLink\">%s</span> from <span class=\"artistLink\">%s</span>", html.EscapeString(song.Name), html.EscapeString(artist.Name)))
			_, cliErr := c.app.restClient.DeleteSong(song.Id)
			if cliErr != nil {
				c.app.HomeComponent.MessageComponent.ClientErrorMessage(fmt.Sprintf("Unable to delete <span class=\"songLink\">%s</span>", html.EscapeString(song.Name)), cliErr)
				return
			}
		}
		c.app.ShowLoader(fmt.Sprintf("Deleting <span class=\"artistLink\">%s</span>", html.EscapeString(artist.Name)))
		_, cliErr := c.app.restClient.DeleteArtist(v)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage(fmt.Sprintf("Unable to delete <span class=\"artistLink\">%s</span>", html.EscapeString(artist.Name)), cliErr)
			return
		}
	case restApiV1.AlbumId:
		album := c.app.localDb.Albums[v]
		for _, song := range c.app.localDb.AlbumOrderedSongs[v] {
			c.app.ShowLoader(fmt.Sprintf("Deleting <span class=\"songLink\">%s</span> from <span class=\"albumLink\">%s</span>", html.EscapeString(song.Name), html.EscapeString(album.Name)))
			_, cliErr := c.app.restClient.DeleteSong(song.Id)
			if cliErr != nil {
				c.app.HomeComponent.MessageComponent.ClientErrorMessage(fmt.Sprintf("Unable to delete <span class=\"songLink\">%s</span>", html.EscapeString(song.Name)), cliErr)
				return
			}
		}
		c.app.ShowLoader(fmt.Sprintf("Deleting <span class=\"albumLink\">%s</span>", html.EscapeString(album.Name)))
		_, cliErr := c.app.restClient.DeleteAlbum(v)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage(fmt.Sprintf("Unable to delete <span class=\"albumLink\">%s</span>", html.EscapeString(album.Name)), cliErr)
			return
		}
	case restApiV1.SongId:
		song := c.app.localDb.Songs[v]
		c.app.ShowLoader(fmt.Sprintf("Deleting <span class=\"songLink\">%s</span>", html.EscapeString(song.Name)))
		_, cliErr := c.app.restClient.DeleteSong(v)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage(fmt.Sprintf("Unable to delete <span class=\"songLink\">%s</span>", html.EscapeString(song.Name)), cliErr)
			return
		}
	case restApiV1.PlaylistId:
		playlist := c.app.localDb.Playlists[v]
		c.app.ShowLoader(fmt.Sprintf("Deleting <span class=\"playlistLink\">%s</span>", html.EscapeString(playlist.Name)))
		_, cliErr := c.app.restClient.DeletePlaylist(v)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage(fmt.Sprintf("Unable to delete <span class=\"playlistLink\">%s</span>", html.EscapeString(playlist.Name)), cliErr)
			return
		}
	case restApiV1.UserId:
		user := c.app.localDb.Users[v]
		c.app.ShowLoader(fmt.Sprintf("Deleting <span class=\"userLink\">%s</span>", html.EscapeString(user.Name)))
		_, cliErr := c.app.restClient.DeleteUser(v)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage(fmt.Sprintf("Unable to delete <span class=\"userLink\">%s</span>", html.EscapeString(user.Name)), cliErr)
			return
		}
	}

	c.app.HomeComponent.Reload()
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
