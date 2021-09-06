package cliwa

import (
	"github.com/jypelle/mifasol/restApiV1"
	"html"
	"strings"
)

type LibraryComponent struct {
	app          *App
	libraryState libraryState
}

type libraryState struct {
	libraryType libraryType
	artistId    *restApiV1.ArtistId
	albumId     *restApiV1.AlbumId
	playlistId  *restApiV1.PlaylistId
	userId      *restApiV1.UserId
	nameFilter  *string
}

type libraryType int64

const (
	libraryTypeArtists libraryType = iota
	libraryTypeAlbums
	libraryTypePlaylists
	libraryTypeSongs
	libraryTypeUsers
)

func NewLibraryComponent(app *App) *LibraryComponent {
	c := &LibraryComponent{
		app: app,
		libraryState: libraryState{
			libraryType: libraryTypeArtists,
			artistId:    nil,
			albumId:     nil,
			playlistId:  nil,
			userId:      nil,
			nameFilter:  nil,
		},
	}

	return c
}

func (c *LibraryComponent) RefreshView() {
	switch c.libraryState.libraryType {
	case libraryTypeArtists:
		c.showLibraryArtistsComponent()
	case libraryTypeAlbums:
		c.showLibraryAlbumsComponent()
	case libraryTypePlaylists:
		c.showLibraryPlaylistsComponent()
	case libraryTypeSongs:
		c.showLibrarySongsComponent()
	}
}

func (c *LibraryComponent) showLibraryArtistsComponent() {
	listDiv := c.app.doc.Call("getElementById", "libraryList")

	var divContent strings.Builder
	for _, artist := range c.app.localDb.OrderedArtists {
		if artist == nil {
			divContent.WriteString(`<div class="artistItem"><a class="artistLink" href="#" onclick="openArtistAction(this.getAttribute('data-artistId'));return false;" data-artistId="` + string(restApiV1.UnknownArtistId) + `">(Unknown artist)</a></div>`)
		} else {
			divContent.WriteString(`<div class="artistItem"><a class="artistLink" href="#" onclick="openArtistAction(this.getAttribute('data-artistId'));return false;" data-artistId="` + string(artist.Id) + `">` + html.EscapeString(artist.Name) + `</a></div>`)
		}
	}
	listDiv.Set("innerHTML", divContent.String())
}

func (c *LibraryComponent) showLibraryAlbumsComponent() {
	listDiv := c.app.doc.Call("getElementById", "libraryList")

	var divContent strings.Builder

	for _, album := range c.app.localDb.OrderedAlbums {
		if album == nil {
			divContent.WriteString(`<div class="albumItem"><a class="albumLink" href="#" onclick="openAlbumAction(this.getAttribute('data-albumId'));return false;" data-albumId="` + string(restApiV1.UnknownAlbumId) + `">(Unknown album)</a>`)
		} else {
			divContent.WriteString(`<div class="albumItem"><a class="albumLink" href="#" onclick="openAlbumAction(this.getAttribute('data-albumId'));return false;" data-albumId="` + string(album.Id) + `">` + html.EscapeString(album.Name) + `</a>`)

			if len(album.ArtistIds) > 0 {
				divContent.WriteString(`<div>`)
				for idx, artistId := range album.ArtistIds {
					if idx > 0 {
						divContent.WriteString(` / `)
					}
					divContent.WriteString(`<a class="artistLink" href="#" onclick="openArtistAction(this.getAttribute('data-artistId'));return false;" data-artistId="` + string(artistId) + `">` + html.EscapeString(c.app.localDb.Artists[artistId].Name) + `</a>`)
				}
				divContent.WriteString(`</div>`)

			}
		}
		divContent.WriteString(`</div>`)
	}
	listDiv.Set("innerHTML", divContent.String())
}

func (c *LibraryComponent) showLibrarySongsComponent() {
	listDiv := c.app.doc.Call("getElementById", "libraryList")

	listDiv.Set("innerHTML", "Loading...")

	var songList []*restApiV1.Song

	if c.libraryState.playlistId == nil {
		if c.libraryState.artistId != nil {
			if *c.libraryState.artistId == restApiV1.UnknownArtistId {
				songList = c.app.localDb.UnknownArtistSongs
			} else {
				songList = c.app.localDb.ArtistOrderedSongs[*c.libraryState.artistId]
			}
		} else if c.libraryState.albumId != nil {
			if *c.libraryState.albumId == restApiV1.UnknownAlbumId {
				songList = c.app.localDb.UnknownAlbumSongs
			} else {
				songList = c.app.localDb.AlbumOrderedSongs[*c.libraryState.albumId]
			}
		} else {
			songList = c.app.localDb.OrderedSongs
		}

		listDiv.Set("innerHTML", "")
		for _, song := range songList {
			listDiv.Call("insertAdjacentHTML", "beforeEnd", c.showSongItem(song))
		}

	} else {
		listDiv.Set("innerHTML", "")
		for _, songId := range c.app.localDb.Playlists[*c.libraryState.playlistId].SongIds {
			listDiv.Call("insertAdjacentHTML", "beforeEnd", c.showSongItem(c.app.localDb.Songs[songId]))
		}
	}

}

func (c *LibraryComponent) showSongItem(song *restApiV1.Song) string {
	var divContent strings.Builder
	divContent.WriteString(`<div class="songItem"><div><a class="songFavoriteLink" href="#" data-songId="` + string(song.Id) + `">`)
	if _, ok := c.app.localDb.UserFavoriteSongIds[c.app.restClient.UserId()][song.Id]; ok {
		divContent.WriteString(`<i class="fas fa-star"></i>`)
	} else {
		divContent.WriteString(`<i class="far fa-star" style="color: #444;"></i>`)
	}
	divContent.WriteString(`</a></div><div><a class="songLink" href="#" onclick="playSongAction(this.getAttribute('data-songId'));return false;" data-songId="` + string(song.Id) + `">` + html.EscapeString(song.Name) + `</a>`)

	if song.AlbumId != restApiV1.UnknownAlbumId || len(song.ArtistIds) > 0 {
		divContent.WriteString(`<div>`)
		if song.AlbumId != restApiV1.UnknownAlbumId {
			divContent.WriteString(`<a class="albumLink" href="#" onclick="openAlbumAction(this.getAttribute('data-albumId'));return false;" data-albumId="` + string(song.AlbumId) + `">` + html.EscapeString(c.app.localDb.Albums[song.AlbumId].Name) + `</a>`)
		} else {
			divContent.WriteString(`<a class="albumLink" href="#" onclick="openAlbumAction(this.getAttribute('data-albumId'));return false;" data-albumId="` + string(song.AlbumId) + `">(Unknown album)</a>`)
		}

		if len(song.ArtistIds) > 0 {
			for _, artistId := range song.ArtistIds {
				divContent.WriteString(` / <a class="artistLink" href="#" onclick="openArtistAction(this.getAttribute('data-artistId'));return false;" data-artistId="` + string(artistId) + `">` + html.EscapeString(c.app.localDb.Artists[artistId].Name) + `</a>`)
			}
		}
		divContent.WriteString(`</div>`)
	}

	divContent.WriteString(`</div></div>`)

	return divContent.String()
}

func (c *LibraryComponent) showLibraryPlaylistsComponent() {
	listDiv := c.app.doc.Call("getElementById", "libraryList")

	var divContent string
	for _, playlist := range c.app.localDb.OrderedPlaylists {
		if playlist != nil {
			divContent += `<div class="playlistItem"><a class="playlistLink" href="#" onclick="openPlaylistAction(this.getAttribute('data-playlistId'));return false;" data-playlistId="` + string(playlist.Id) + `">` + html.EscapeString(playlist.Name) + `</a></div>`
		}
	}
	listDiv.Set("innerHTML", divContent)
}
