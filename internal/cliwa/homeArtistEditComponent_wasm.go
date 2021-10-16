package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restApiV1"
)

type HomeArtistEditComponent struct {
	app        *App
	artistId   restApiV1.ArtistId
	artistMeta *restApiV1.ArtistMeta
	closed     bool
}

func NewHomeArtistEditComponent(app *App, artistId restApiV1.ArtistId, artistMeta restApiV1.ArtistMeta) *HomeArtistEditComponent {
	c := &HomeArtistEditComponent{
		app:        app,
		artistId:   artistId,
		artistMeta: &artistMeta,
	}

	return c
}

func (c *HomeArtistEditComponent) Show() {
	div := jst.Document.Call("getElementById", "homeMainModal")
	div.Set("innerHTML", c.app.RenderTemplate(
		c.artistMeta, "home/artistEdit/index"),
	)

	form := jst.Document.Call("getElementById", "artistEditForm")
	form.Call("addEventListener", "submit", c.app.AddEventFuncPreventDefault(c.saveAction))
	cancelButton := jst.Document.Call("getElementById", "artistEditCancelButton")
	cancelButton.Call("addEventListener", "click", c.app.AddEventFunc(c.cancelAction))

}

func (c *HomeArtistEditComponent) saveAction() {
	if c.closed {
		return
	}

	c.app.ShowLoader("Updating all songs of the artist")

	artistName := jst.Document.Call("getElementById", "artistEditArtistName")
	c.artistMeta.Name = artistName.Get("value").String()

	if c.artistId != "" {
		_, cliErr := c.app.restClient.UpdateArtist(c.artistId, c.artistMeta)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to update the artist", cliErr)
		}
	} else {
		_, cliErr := c.app.restClient.CreateArtist(c.artistMeta)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to create the artist", cliErr)
		}
	}
	c.close()
	c.app.HomeComponent.Reload()
	c.app.HideLoader()
}

func (c *HomeArtistEditComponent) cancelAction() {
	if c.closed {
		return
	}
	c.close()
}

func (c *HomeArtistEditComponent) close() {
	c.closed = true
	c.app.HomeComponent.CloseModal()
}
