package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restApiV1"
)

type HomeSongEditComponent struct {
	app      *App
	songId   restApiV1.SongId
	songMeta *restApiV1.SongMeta
	closed   bool
}

func NewHomeSongEditComponent(app *App, songId restApiV1.SongId, songMeta *restApiV1.SongMeta) *HomeSongEditComponent {
	c := &HomeSongEditComponent{
		app:      app,
		songId:   songId,
		songMeta: songMeta.Copy(),
	}

	return c
}

func (c *HomeSongEditComponent) Show() {
	div := jst.Document.Call("getElementById", "homeMainModal")
	div.Set("innerHTML", c.app.RenderTemplate(
		c.songMeta, "home/songEdit/index"),
	)

	form := jst.Document.Call("getElementById", "songEditForm")
	form.Call("addEventListener", "submit", c.app.AddEventFuncPreventDefault(c.saveAction))
	cancelButton := jst.Document.Call("getElementById", "songEditCancelButton")
	cancelButton.Call("addEventListener", "click", c.app.AddEventFunc(c.cancelAction))

}

func (c *HomeSongEditComponent) saveAction() {
	if c.closed {
		return
	}

	c.app.ShowLoader("Updating song")

	songName := jst.Document.Call("getElementById", "songEditSongName")
	c.songMeta.Name = songName.Get("value").String()

	_, cliErr := c.app.restClient.UpdateSong(c.songId, c.songMeta)
	if cliErr != nil {
		c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to update the song", cliErr)
	}

	c.close()
	c.app.HomeComponent.Reload()
	c.app.HideLoader()
}

func (c *HomeSongEditComponent) cancelAction() {
	if c.closed {
		return
	}
	c.close()
}

func (c *HomeSongEditComponent) close() {
	c.closed = true
	c.app.HomeComponent.CloseModal()
}
