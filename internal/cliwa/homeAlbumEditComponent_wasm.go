package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restApiV1"
)

type HomeAlbumEditComponent struct {
	app       *App
	albumId   restApiV1.AlbumId
	albumMeta *restApiV1.AlbumMeta
	closed    bool
}

func NewHomeAlbumEditComponent(app *App, albumId restApiV1.AlbumId, albumMeta *restApiV1.AlbumMeta) *HomeAlbumEditComponent {
	c := &HomeAlbumEditComponent{
		app:       app,
		albumId:   albumId,
		albumMeta: albumMeta.Copy(),
	}

	return c
}

func (c *HomeAlbumEditComponent) Render() {
	div := jst.Id("homeMainModal")
	div.Set("innerHTML", c.app.RenderTemplate(
		c.albumMeta, "home/albumEdit/index"),
	)

	form := jst.Id("albumEditForm")
	form.Call("addEventListener", "submit", c.app.AddEventFuncPreventDefault(c.saveAction))
	cancelButton := jst.Id("albumEditCancelButton")
	cancelButton.Call("addEventListener", "click", c.app.AddEventFunc(c.cancelAction))

}

func (c *HomeAlbumEditComponent) saveAction() {
	if c.closed {
		return
	}

	c.app.ShowLoader("Updating all songs of the album")

	albumName := jst.Id("albumEditAlbumName")
	c.albumMeta.Name = albumName.Get("value").String()

	if c.albumId != "" {
		_, cliErr := c.app.restClient.UpdateAlbum(c.albumId, c.albumMeta)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to update the album", cliErr)
		}
	} else {
		_, cliErr := c.app.restClient.CreateAlbum(c.albumMeta)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to create the album", cliErr)
		}
	}
	c.close()
	c.app.HomeComponent.Reload()
	c.app.HideLoader()
}

func (c *HomeAlbumEditComponent) cancelAction() {
	if c.closed {
		return
	}
	c.close()
}

func (c *HomeAlbumEditComponent) close() {
	c.closed = true
	c.app.HomeComponent.CloseModal()
}
