package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restApiV1"
)

type HomeAlbumEditComponent struct {
	app       *App
	albumId   restApiV1.AlbumId
	albumMeta *restApiV1.AlbumMeta
}

func NewHomeAlbumEditComponent(app *App, albumId restApiV1.AlbumId, albumMeta *restApiV1.AlbumMeta) *HomeAlbumEditComponent {
	c := &HomeAlbumEditComponent{
		app:       app,
		albumId:   albumId,
		albumMeta: albumMeta,
	}

	return c
}

func (c *HomeAlbumEditComponent) Show() {
	div := jst.Document.Call("getElementById", "homeMainModal")
	div.Set("innerHTML", c.app.RenderTemplate(
		c.albumMeta, "homeAlbumEditComponent.html"),
	)

	form := jst.Document.Call("getElementById", "albumEditForm")
	form.Call("addEventListener", "submit", c.app.AddEventFuncPreventDefault(c.saveAction))
	cancelButton := jst.Document.Call("getElementById", "albumEditCancelButton")
	cancelButton.Call("addEventListener", "click", c.app.AddEventFunc(c.cancelAction))

}

func (c *HomeAlbumEditComponent) saveAction() {
	c.close()
}

func (c *HomeAlbumEditComponent) cancelAction() {
	c.close()
}

func (c *HomeAlbumEditComponent) close() {
	homeMainMaster := jst.Document.Call("getElementById", "homeMainMaster")
	homeMainMaster.Get("style").Set("display", "flex")
	homeMainModal := jst.Document.Call("getElementById", "homeMainModal")
	homeMainModal.Set("innerHTML", "")
	homeMainModal.Get("style").Set("display", "none")
}
