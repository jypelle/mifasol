package cliwa

import (
	"html"
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

	go func() {
		c.Refresh()
		c.showLibraryArtistsComponent()
	}()

}

func (c *App) showLibraryArtistsComponent() {
	listDiv := c.doc.Call("getElementById", "libraryList")

	var divContent string
	for _, artist := range c.localDb.OrderedArtists {
		if artist != nil {
			divContent += "<p>" + html.EscapeString(artist.Name) + "</p>"
		}
	}
	listDiv.Set("innerHTML", divContent)
}

func (c *App) showLibraryAlbumsComponent() {
	listDiv := c.doc.Call("getElementById", "libraryList")

	var divContent string
	for _, album := range c.localDb.OrderedAlbums {
		if album != nil {
			divContent += "<p>" + html.EscapeString(album.Name) + "</p>"
		}
	}
	listDiv.Set("innerHTML", divContent)
}

func (c *App) showLibrarySongsComponent() {
	listDiv := c.doc.Call("getElementById", "libraryList")

	var divContent string
	for _, song := range c.localDb.OrderedSongs {
		if song != nil {
			divContent += "<p>" + html.EscapeString(song.Name) + "</p>"
		}
	}
	listDiv.Set("innerHTML", divContent)
}

func (c *App) showLibraryPlaylistsComponent() {
	listDiv := c.doc.Call("getElementById", "libraryList")

	var divContent string
	for _, playlist := range c.localDb.OrderedPlaylists {
		if playlist != nil {
			divContent += "<p>" + html.EscapeString(playlist.Name) + "</p>"
		}
	}
	listDiv.Set("innerHTML", divContent)
}
