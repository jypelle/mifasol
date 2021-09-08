package cliwa

import (
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/sirupsen/logrus"
	"html"
	"strings"
	"syscall/js"
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

func (c *LibraryComponent) Show() {
	libraryArtistsButton := c.app.doc.Call("getElementById", "libraryArtistsButton")
	libraryArtistsButton.Call("addEventListener", "click", js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		c.ShowArtistsAction()
		return nil
	}))
	libraryAlbumsButton := c.app.doc.Call("getElementById", "libraryAlbumsButton")
	libraryAlbumsButton.Call("addEventListener", "click", js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		c.ShowAlbumsAction()
		return nil
	}))
	librarySongsButton := c.app.doc.Call("getElementById", "librarySongsButton")
	librarySongsButton.Call("addEventListener", "click", js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		c.ShowSongsAction()
		return nil
	}))
	libraryPlaylistsButton := c.app.doc.Call("getElementById", "libraryPlaylistsButton")
	libraryPlaylistsButton.Call("addEventListener", "click", js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		c.ShowPlaylistsAction()
		return nil
	}))

	listDiv := c.app.doc.Call("getElementById", "libraryList")
	listDiv.Call("addEventListener", "click", js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		link := i[0].Get("target").Call("closest", ".songLink, .artistLink, .albumLink, .playlistLink, .songFavoriteLink")
		if !link.Truthy() {
			return nil
		}
		dataset := link.Get("dataset")

		switch link.Get("className").String() {
		case "songLink":
			songId := dataset.Get("songid").String()
			logrus.Infof("click on %v - %v", dataset, songId)
			c.app.playSong(restApiV1.SongId(songId))
		case "artistLink":
			artistId := dataset.Get("artistid").String()
			c.OpenArtistAction(restApiV1.ArtistId(artistId))
		case "albumLink":
			albumId := dataset.Get("albumid").String()
			c.OpenAlbumAction(restApiV1.AlbumId(albumId))
		case "playlistLink":
			playlistId := dataset.Get("playlistid").String()
			c.OpenPlaylistAction(restApiV1.PlaylistId(playlistId))
		case "songFavoriteLink":
			songId := dataset.Get("songid").String()
			favoriteSongId := restApiV1.FavoriteSongId{
				UserId: c.app.restClient.UserId(),
				SongId: restApiV1.SongId(songId),
			}

			go func() {

				if _, ok := c.app.localDb.UserFavoriteSongIds[c.app.restClient.UserId()][restApiV1.SongId(songId)]; ok {
					link.Set("innerHTML", `<i class="far fa-star" style="color: #444;"></i>`)

					_, cliErr := c.app.restClient.DeleteFavoriteSong(favoriteSongId)
					if cliErr != nil {
						c.app.messageComponent.Message("Unable to add song to favorites")
						link.Set("innerHTML", `<i class="fas fa-star"></i>`)
						return
					}
					c.app.localDb.RemoveSongFromMyFavorite(restApiV1.SongId(songId))

					logrus.Info("Deactivate")
				} else {
					link.Set("innerHTML", `<i class="fas fa-star"></i>`)

					_, cliErr := c.app.restClient.CreateFavoriteSong(&restApiV1.FavoriteSongMeta{Id: favoriteSongId})
					if cliErr != nil {
						c.app.messageComponent.Message("Unable to remove song from favorites")
						link.Set("innerHTML", `<i class="far fa-star" style="color: #444;"></i>`)
						return
					}
					c.app.localDb.AddSongToMyFavorite(restApiV1.SongId(songId))

					logrus.Info("Activate")
				}
			}()

		}

		return nil
	}))

}

func (c *LibraryComponent) RefreshView() {
	switch c.libraryState.libraryType {
	case libraryTypeArtists:
		c.refreshArtists()
	case libraryTypeAlbums:
		c.refreshAlbums()
	case libraryTypePlaylists:
		c.refreshPlaylists()
	case libraryTypeSongs:
		c.refreshSongs()
	}
}

func (c *LibraryComponent) refreshArtists() {
	listDiv := c.app.doc.Call("getElementById", "libraryList")

	var divContent strings.Builder
	for _, artist := range c.app.localDb.OrderedArtists {
		if artist == nil {
			divContent.WriteString(`<div class="artistItem"><a class="artistLink" href="#" data-artistid="` + string(restApiV1.UnknownArtistId) + `">(Unknown artist)</a></div>`)
		} else {
			divContent.WriteString(`<div class="artistItem"><a class="artistLink" href="#" data-artistid="` + string(artist.Id) + `">` + html.EscapeString(artist.Name) + `</a></div>`)
		}
	}
	listDiv.Set("innerHTML", divContent.String())
}

func (c *LibraryComponent) refreshAlbums() {
	listDiv := c.app.doc.Call("getElementById", "libraryList")

	var divContent strings.Builder

	for _, album := range c.app.localDb.OrderedAlbums {
		if album == nil {
			divContent.WriteString(`<div class="albumItem"><a class="albumLink" href="#" data-albumid="` + string(restApiV1.UnknownAlbumId) + `">(Unknown album)</a>`)
		} else {
			divContent.WriteString(`<div class="albumItem"><a class="albumLink" href="#" data-albumid="` + string(album.Id) + `">` + html.EscapeString(album.Name) + `</a>`)

			if len(album.ArtistIds) > 0 {
				divContent.WriteString(`<div>`)
				for idx, artistId := range album.ArtistIds {
					if idx > 0 {
						divContent.WriteString(` / `)
					}
					divContent.WriteString(`<a class="artistLink" href="#" data-artistid="` + string(artistId) + `">` + html.EscapeString(c.app.localDb.Artists[artistId].Name) + `</a>`)
				}
				divContent.WriteString(`</div>`)

			}
		}
		divContent.WriteString(`</div>`)
	}
	listDiv.Set("innerHTML", divContent.String())
}

func (c *LibraryComponent) refreshSongs() {

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
			listDiv.Call("insertAdjacentHTML", "beforeEnd", c.addSongItem(song))
		}

	} else {
		listDiv.Set("innerHTML", "")
		for _, songId := range c.app.localDb.Playlists[*c.libraryState.playlistId].SongIds {
			listDiv.Call("insertAdjacentHTML", "beforeEnd", c.addSongItem(c.app.localDb.Songs[songId]))
		}
	}
}

func (c *LibraryComponent) addSongItem(song *restApiV1.Song) string {
	var divContent strings.Builder
	divContent.WriteString(`<div class="songItem"><div><a class="songFavoriteLink" href="#" data-songid="` + string(song.Id) + `">`)
	if _, ok := c.app.localDb.UserFavoriteSongIds[c.app.restClient.UserId()][song.Id]; ok {
		divContent.WriteString(`<i class="fas fa-star"></i>`)
	} else {
		divContent.WriteString(`<i class="far fa-star" style="color: #444;"></i>`)
	}
	divContent.WriteString(`</a></div><div><a class="songLink" href="#" data-songid="` + string(song.Id) + `">` + html.EscapeString(song.Name) + `</a>`)

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

	divContent.WriteString(`</div></div>`)

	return divContent.String()
}

func (c *LibraryComponent) refreshPlaylists() {
	listDiv := c.app.doc.Call("getElementById", "libraryList")

	var divContent string
	for _, playlist := range c.app.localDb.OrderedPlaylists {
		if playlist != nil {
			divContent += `<div class="playlistItem"><a class="playlistLink" href="#" data-playlistid="` + string(playlist.Id) + `">` + html.EscapeString(playlist.Name) + `</a></div>`
		}
	}
	listDiv.Set("innerHTML", divContent)
}

func (c *LibraryComponent) ShowArtistsAction() {
	c.libraryState = libraryState{
		libraryType: libraryTypeArtists,
	}
	c.RefreshView()
}

func (c *LibraryComponent) ShowAlbumsAction() {
	c.libraryState = libraryState{
		libraryType: libraryTypeAlbums,
	}
	c.RefreshView()
}

func (c *LibraryComponent) ShowSongsAction() {
	c.libraryState = libraryState{
		libraryType: libraryTypeSongs,
	}
	c.RefreshView()
}

func (c *LibraryComponent) ShowPlaylistsAction() {
	c.libraryState = libraryState{
		libraryType: libraryTypePlaylists,
	}
	c.RefreshView()
}

func (c *LibraryComponent) OpenAlbumAction(albumId restApiV1.AlbumId) {
	c.libraryState = libraryState{
		libraryType: libraryTypeSongs,
		albumId:     &albumId,
	}
	c.RefreshView()
}

func (c *LibraryComponent) OpenArtistAction(artistId restApiV1.ArtistId) {
	c.libraryState = libraryState{
		libraryType: libraryTypeSongs,
		artistId:    &artistId,
	}
	c.RefreshView()
}

func (c *LibraryComponent) OpenPlaylistAction(playlistId restApiV1.PlaylistId) {
	c.libraryState = libraryState{
		libraryType: libraryTypeSongs,
		playlistId:  &playlistId,
	}
	c.RefreshView()
}
