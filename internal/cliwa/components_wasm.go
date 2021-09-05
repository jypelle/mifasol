package cliwa

import (
	"net/url"
	"strconv"
	"syscall/js"
)

func (c *App) retrieveServerCredentials() {
	rawUrl := js.Global().Get("window").Get("location").Get("href").String()
	baseUrl, _ := url.Parse(rawUrl)

	c.config.ServerHostname = baseUrl.Hostname()
	c.config.ServerPort, _ = strconv.ParseInt(baseUrl.Port(), 10, 64)
	c.config.ServerSsl = baseUrl.Scheme == "https"
}

func (c *App) showStartComponent() {
	c.restClient = nil
	c.localDb = nil

	body := c.doc.Get("body")
	body.Set("innerHTML", c.RenderTemplate(nil, "start.html"))

	// Set focus
	c.doc.Call("getElementById", "mifasolUsername").Call("focus")

	// Set button
	js.Global().Set("logInAction", js.FuncOf(c.logInAction))
}

func (c *App) showHomeComponent() {
	body := c.doc.Get("body")
	body.Set("innerHTML", c.RenderTemplate(nil, "home.html"))

	// Set callback
	js.Global().Set("logOutAction", js.FuncOf(c.logOutAction))
	js.Global().Set("refreshAction", js.FuncOf(c.refreshAction))
	js.Global().Set("showLibraryArtistsAction", js.FuncOf(c.showLibraryArtistsAction))
	js.Global().Set("showLibraryAlbumsAction", js.FuncOf(c.showLibraryAlbumsAction))
	js.Global().Set("showLibrarySongsAction", js.FuncOf(c.showLibrarySongsAction))
	js.Global().Set("showLibraryPlaylistsAction", js.FuncOf(c.showLibraryPlaylistsAction))
	js.Global().Set("playSongAction", js.FuncOf(c.playSongAction))
	js.Global().Set("openArtistAction", js.FuncOf(c.openArtistAction))
	js.Global().Set("openAlbumAction", js.FuncOf(c.openAlbumAction))
	js.Global().Set("openPlaylistAction", js.FuncOf(c.openPlaylistAction))

	go func() {
		c.Reload()
	}()

}
