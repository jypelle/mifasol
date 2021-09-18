package cliwa

import (
	"github.com/jypelle/mifasol/restApiV1"
	"html"
	"strings"
	"syscall/js"
)

type CurrentComponent struct {
	app *App

	songIds       []restApiV1.SongId
	srcPlaylistId *restApiV1.PlaylistId
	modified      bool
}

func NewCurrentComponent(app *App) *CurrentComponent {
	c := &CurrentComponent{
		app: app,
	}

	return c
}

func (c *CurrentComponent) Show() {
	currentClearButton := c.app.doc.Call("getElementById", "currentClearButton")
	currentClearButton.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		c.songIds = nil
		c.RefreshView()
	}))

	listDiv := c.app.doc.Call("getElementById", "currentList")
	listDiv.Call("addEventListener", "click", c.app.AddRichEventFunc(func(this js.Value, i []js.Value) {
		link := i[0].Get("target").Call("closest", ".artistLink, .albumLink, .songPlayNowLink")
		if !link.Truthy() {
			return
		}
		dataset := link.Get("dataset")

		switch link.Get("className").String() {
		case "artistLink":
			artistId := dataset.Get("artistid").String()
			c.app.libraryComponent.OpenArtistAction(restApiV1.ArtistId(artistId))
		case "albumLink":
			albumId := dataset.Get("albumid").String()
			c.app.libraryComponent.OpenAlbumAction(restApiV1.AlbumId(albumId))
		case "songPlayNowLink":
			songId := dataset.Get("songid").String()
			c.app.playSong(restApiV1.SongId(songId))
		}
	}))

}

func (c *CurrentComponent) RefreshView() {
	listDiv := c.app.doc.Call("getElementById", "currentList")

	listDiv.Set("innerHTML", "")
	for _, songId := range c.songIds {
		listDiv.Call("insertAdjacentHTML", "beforeEnd", c.addSongItem(c.app.localDb.Songs[songId]))
	}
}

func (c *CurrentComponent) addSongItem(song *restApiV1.Song) string {
	var divContent strings.Builder
	divContent.WriteString(`<div class="item">`)

	// Title item
	divContent.WriteString(`<div class="itemTitle"><p class="songLink">` + html.EscapeString(song.Name) + `</p>`)
	if song.AlbumId != restApiV1.UnknownAlbumId || len(song.ArtistIds) > 0 {
		divContent.WriteString(`<div>`)
		if song.AlbumId != restApiV1.UnknownAlbumId {
			divContent.WriteString(`<a class="albumLink" href="#" data-albumid="` + string(song.AlbumId) + `">` + html.EscapeString(c.app.localDb.Albums[song.AlbumId].Name) + `</a>`)
		} else {
			divContent.WriteString(`<a class="albumLink" href="#" data-albumid="` + string(song.AlbumId) + `">(Unknown album)</a>`)
		}

		if len(song.ArtistIds) > 0 {
			for _, artistId := range song.ArtistIds {
				divContent.WriteString(` / <a class="artistLink" href="#" data-artistid="` + string(artistId) + `">` + html.EscapeString(c.app.localDb.Artists[artistId].Name) + `</a>`)
			}
		}
		divContent.WriteString(`</div>`)
	}
	divContent.WriteString(`</div>`)

	// Buttons item
	divContent.WriteString(`<div class="itemButtons">`)

	// 'Play now' button
	divContent.WriteString(`<a class="songPlayNowLink" href="#" data-songid="` + string(song.Id) + `">`)
	divContent.WriteString(`<i class="fas fa-play"></i>`)
	divContent.WriteString(`</a>`)

	divContent.WriteString(`</div>`)

	divContent.WriteString(`</div>`)

	return divContent.String()
}

func (c *CurrentComponent) AddSongAction(songId restApiV1.SongId) {
	c.songIds = append(c.songIds, songId)
	c.RefreshView()
}

func (c *CurrentComponent) AddSongsFromAlbumAction(albumId restApiV1.AlbumId) {
	if albumId != restApiV1.UnknownAlbumId {
		for _, song := range c.app.localDb.AlbumOrderedSongs[albumId] {
			c.AddSongAction(song.Id)
		}
	} else {
		for _, song := range c.app.localDb.UnknownAlbumSongs {
			c.AddSongAction(song.Id)
		}
	}
}

func (c *CurrentComponent) AddSongsFromArtistAction(artistId restApiV1.ArtistId) {
	if artistId != restApiV1.UnknownArtistId {
		for _, song := range c.app.localDb.ArtistOrderedSongs[artistId] {
			c.AddSongAction(song.Id)
		}
	} else {
		for _, song := range c.app.localDb.UnknownArtistSongs {
			c.AddSongAction(song.Id)
		}
	}
}

func (c *CurrentComponent) AddSongsFromPlaylistAction(playlistId restApiV1.PlaylistId) {
	for _, songId := range c.app.localDb.Playlists[playlistId].SongIds {
		c.AddSongAction(songId)
	}
}
