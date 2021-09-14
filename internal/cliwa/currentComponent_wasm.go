package cliwa

import (
	"github.com/jypelle/mifasol/restApiV1"
	"html"
	"strings"
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

func (c *CurrentComponent) RefreshView() {
	listDiv := c.app.doc.Call("getElementById", "currentList")

	listDiv.Set("innerHTML", "")
	for _, songId := range c.songIds {
		listDiv.Call("insertAdjacentHTML", "beforeEnd", c.addSongItem(c.app.localDb.Songs[songId]))
	}
}

func (c *CurrentComponent) addSongItem(song *restApiV1.Song) string {
	var divContent strings.Builder
	divContent.WriteString(`<div class="songItem">`)

	divContent.WriteString(`<div></div>`)

	// Song block
	divContent.WriteString(`<div><a class="songLink" href="#" data-songid="` + string(song.Id) + `">` + html.EscapeString(song.Name) + `</a>`)
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
