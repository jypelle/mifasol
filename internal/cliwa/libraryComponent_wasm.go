package cliwa

import (
	"fmt"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/sirupsen/logrus"
	"html"
	"strings"
	"syscall/js"
)

type libraryType int64

const (
	libraryTypeArtists libraryType = iota
	libraryTypeAlbums
	libraryTypePlaylists
	libraryTypeSongs
	libraryTypeUsers
)

type libraryState struct {
	libraryType libraryType
	artistId    *restApiV1.ArtistId
	albumId     *restApiV1.AlbumId
	playlistId  *restApiV1.PlaylistId
	userId      *restApiV1.UserId
	nameFilter  *string
}

func (c *LibraryComponent) refreshTitle() {

	var title string

	switch c.libraryState.libraryType {
	case libraryTypeArtists:
		if c.libraryState.userId == nil {
			title = "Artists"
		} else {
			title = fmt.Sprintf("Favorite artists from %s", c.app.localDb.Users[*c.libraryState.userId].Name)
		}
	case libraryTypeAlbums:
		if c.libraryState.userId == nil {
			title = "Albums"
		} else {
			title = fmt.Sprintf("Favorite albums from %s", c.app.localDb.Users[*c.libraryState.userId].Name)
		}
	case libraryTypePlaylists:
		if c.libraryState.userId == nil {
			title = "Playlists"
		} else {
			title = fmt.Sprintf("Favorite playlists from %s", c.app.localDb.Users[*c.libraryState.userId].Name)
		}
	case libraryTypeSongs:
		if c.libraryState.userId == nil && c.libraryState.playlistId == nil && c.libraryState.artistId == nil && c.libraryState.albumId == nil {
			title = "Songs"
		}
		if c.libraryState.playlistId != nil {
			title = fmt.Sprintf("Songs from %s", c.app.localDb.Playlists[*c.libraryState.playlistId].Name)
		}
		if c.libraryState.userId != nil {
			title = fmt.Sprintf("Favorite songs from %s", c.app.localDb.Users[*c.libraryState.userId].Name)
		}
		if c.libraryState.artistId != nil {
			if *c.libraryState.artistId != restApiV1.UnknownArtistId {
				title = fmt.Sprintf("Songs from %s", c.app.localDb.Artists[*c.libraryState.artistId].Name)
			} else {
				title = "Songs from unknown artists"
			}
		}
		if c.libraryState.albumId != nil {
			if *c.libraryState.albumId != restApiV1.UnknownAlbumId {
				title = fmt.Sprintf("Songs from %s", c.app.localDb.Albums[*c.libraryState.albumId].Name)
			} else {
				title = "Songs from unknown albums"
			}
		}
	case libraryTypeUsers:
		title = "Users"
	}

	titleSpan := c.app.doc.Call("getElementById", "libraryTitle")
	titleSpan.Set("innerHTML", title)
}

type LibraryComponent struct {
	app          *App
	libraryState libraryState
}

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
	libraryUsersButton := c.app.doc.Call("getElementById", "libraryUsersButton")
	libraryUsersButton.Call("addEventListener", "click", js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		c.ShowUsersAction()
		return nil
	}))

	listDiv := c.app.doc.Call("getElementById", "libraryList")
	listDiv.Call("addEventListener", "click", js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		link := i[0].Get("target").Call("closest", ".artistLink, .artistAddToPlaylistLink, .albumLink, .albumAddToPlaylistLink, .playlistLink, .playlistFavoriteLink, .playlistAddToPlaylistLink, .songFavoriteLink, .songAddToPlaylistLink, .songPlayNowLink, .songDownloadLink")
		if !link.Truthy() {
			return nil
		}
		dataset := link.Get("dataset")

		switch link.Get("className").String() {
		case "artistLink":
			artistId := dataset.Get("artistid").String()
			c.OpenArtistAction(restApiV1.ArtistId(artistId))
		case "artistAddToPlaylistLink":
			artistId := dataset.Get("artistid").String()
			c.app.currentComponent.AddSongsFromArtistAction(restApiV1.ArtistId(artistId))
		case "albumLink":
			albumId := dataset.Get("albumid").String()
			c.OpenAlbumAction(restApiV1.AlbumId(albumId))
		case "albumAddToPlaylistLink":
			albumId := dataset.Get("albumid").String()
			c.app.currentComponent.AddSongsFromAlbumAction(restApiV1.AlbumId(albumId))
		case "playlistLink":
			playlistId := dataset.Get("playlistid").String()
			c.OpenPlaylistAction(restApiV1.PlaylistId(playlistId))
		case "playlistFavoriteLink":
			playlistId := dataset.Get("playlistid").String()
			favoritePlaylistId := restApiV1.FavoritePlaylistId{
				UserId:     c.app.restClient.UserId(),
				PlaylistId: restApiV1.PlaylistId(playlistId),
			}

			go func() {

				if _, ok := c.app.localDb.UserFavoritePlaylistIds[c.app.restClient.UserId()][restApiV1.PlaylistId(playlistId)]; ok {
					link.Set("innerHTML", `<i class="far fa-star" style="color: #444;"></i>`)

					_, cliErr := c.app.restClient.DeleteFavoritePlaylist(favoritePlaylistId)
					if cliErr != nil {
						c.app.messageComponent.Message("Unable to add playlist to favorites")
						link.Set("innerHTML", `<i class="fas fa-star"></i>`)
						return
					}
					go c.app.Reload()
					//c.app.localDb.RemovePlaylistFromMyFavorite(restApiV1.PlaylistId(playlistId))

				} else {
					link.Set("innerHTML", `<i class="fas fa-star"></i>`)

					_, cliErr := c.app.restClient.CreateFavoritePlaylist(&restApiV1.FavoritePlaylistMeta{Id: favoritePlaylistId})
					if cliErr != nil {
						c.app.messageComponent.Message("Unable to remove playlist from favorites")
						link.Set("innerHTML", `<i class="far fa-star" style="color: #444;"></i>`)
						return
					}
					go c.app.Reload()
					//c.app.localDb.AddPlaylistToMyFavorite(restApiV1.PlaylistId(playlistId))

				}
			}()
		case "playlistAddToPlaylistLink":
			playlistId := dataset.Get("playlistid").String()
			c.app.currentComponent.AddSongsFromPlaylistAction(restApiV1.PlaylistId(playlistId))
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
		case "songPlayNowLink":
			songId := dataset.Get("songid").String()
			//			logrus.Debugf("click on %v - %v", dataset, songId)
			c.app.playSong(restApiV1.SongId(songId))
		case "songDownloadLink":
			songId := dataset.Get("songid").String()
			song := c.app.localDb.Songs[restApiV1.SongId(songId)]

			go func() {
				token, cliErr := c.app.restClient.GetToken()
				if cliErr != nil {
					return
				}

				anchor := c.app.doc.Call("createElement", "a")
				anchor.Set("href", "/api/v1/songContents/"+string(songId)+"?bearer="+token.AccessToken)
				anchor.Set("download", song.Name+song.Format.Extension())
				c.app.doc.Get("body").Call("appendChild", anchor)
				anchor.Call("click")
				c.app.doc.Get("body").Call("removeChild", anchor)
			}()
		case "songAddToPlaylistLink":
			songId := dataset.Get("songid").String()
			c.app.currentComponent.AddSongAction(restApiV1.SongId(songId))
		}

		return nil
	}))

}

func (c *LibraryComponent) RefreshView() {
	// Refresh library list
	switch c.libraryState.libraryType {
	case libraryTypeArtists:
		c.refreshArtistList()
	case libraryTypeAlbums:
		c.refreshAlbumList()
	case libraryTypePlaylists:
		c.refreshPlaylistList()
	case libraryTypeSongs:
		c.refreshSongList()
	case libraryTypeUsers:
		c.refreshUserList()
	}

	// Refresh library title
	c.refreshTitle()
}

func (c *LibraryComponent) refreshArtistList() {
	listDiv := c.app.doc.Call("getElementById", "libraryList")

	var divContent strings.Builder
	for _, artist := range c.app.localDb.OrderedArtists {
		divContent.WriteString(`<div class="item artistItem">`)

		// Title item
		divContent.WriteString(`<div class="itemTitle">`)

		if artist == nil {
			divContent.WriteString(`<a class="artistLink" href="#" data-artistid="` + string(restApiV1.UnknownArtistId) + `">(Unknown artist)</a><div></div>`)
		} else {
			divContent.WriteString(`<a class="artistLink" href="#" data-artistid="` + string(artist.Id) + `">` + html.EscapeString(artist.Name) + `</a><div></div>`)
		}
		divContent.WriteString(`</div>`)

		// Buttons item
		divContent.WriteString(`<div class="itemButtons">`)

		// 'Add to current playlist' button
		if artist == nil {
			divContent.WriteString(`<a class="artistAddToPlaylistLink" href="#" data-artistid="` + string(restApiV1.UnknownArtistId) + `">`)
		} else {
			divContent.WriteString(`<a class="artistAddToPlaylistLink" href="#" data-artistid="` + string(artist.Id) + `">`)
		}
		divContent.WriteString(`<i class="fas fa-plus"></i>`)
		divContent.WriteString(`</a>`)

		divContent.WriteString(`</div>`)

		divContent.WriteString(`</div>`)

	}
	listDiv.Set("innerHTML", divContent.String())
}

func (c *LibraryComponent) refreshAlbumList() {
	listDiv := c.app.doc.Call("getElementById", "libraryList")

	var divContent strings.Builder

	for _, album := range c.app.localDb.OrderedAlbums {
		divContent.WriteString(`<div class="item albumItem">`)

		// Title item
		divContent.WriteString(`<div class="itemTitle">`)
		if album == nil {
			divContent.WriteString(`<a class="albumLink" href="#" data-albumid="` + string(restApiV1.UnknownAlbumId) + `">(Unknown album)</a><div></div>`)
		} else {
			divContent.WriteString(`<a class="albumLink" href="#" data-albumid="` + string(album.Id) + `">` + html.EscapeString(album.Name) + `</a>`)

			divContent.WriteString(`<div>`)
			for idx, artistId := range album.ArtistIds {
				if idx > 0 {
					divContent.WriteString(` / `)
				}
				divContent.WriteString(`<a class="artistLink" href="#" data-artistid="` + string(artistId) + `">` + html.EscapeString(c.app.localDb.Artists[artistId].Name) + `</a>`)
			}
			divContent.WriteString(`</div>`)
		}
		divContent.WriteString(`</div>`)

		// Buttons item
		divContent.WriteString(`<div class="itemButtons">`)

		// 'Add to current playlist' button
		if album == nil {
			divContent.WriteString(`<a class="albumAddToPlaylistLink" href="#" data-albumid="` + string(restApiV1.UnknownAlbumId) + `">`)
		} else {
			divContent.WriteString(`<a class="albumAddToPlaylistLink" href="#" data-albumid="` + string(album.Id) + `">`)
		}
		divContent.WriteString(`<i class="fas fa-plus"></i>`)
		divContent.WriteString(`</a>`)

		divContent.WriteString(`</div>`)

		divContent.WriteString(`</div>`)
	}
	listDiv.Set("innerHTML", divContent.String())
}

func (c *LibraryComponent) refreshSongList() {

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
	divContent.WriteString(`<div class="item songItem">`)

	// Favorite item
	divContent.WriteString(`<div class="itemFavorite"><a class="songFavoriteLink" href="#" data-songid="` + string(song.Id) + `">`)
	if _, ok := c.app.localDb.UserFavoriteSongIds[c.app.restClient.UserId()][song.Id]; ok {
		divContent.WriteString(`<i class="fas fa-star"></i>`)
	} else {
		divContent.WriteString(`<i class="far fa-star" style="color: #444;"></i>`)
	}
	divContent.WriteString(`</a></div>`)

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

	// 'Download song' button
	divContent.WriteString(`<a class="songDownloadLink" href="#" data-songid="` + string(song.Id) + `">`)
	divContent.WriteString(`<i class="fas fa-arrow-down"></i>`)
	divContent.WriteString(`</a>`)

	// 'Play now' button
	divContent.WriteString(`<a class="songPlayNowLink" href="#" data-songid="` + string(song.Id) + `">`)
	divContent.WriteString(`<i class="fas fa-play"></i>`)
	divContent.WriteString(`</a>`)

	// 'Add to current playlist' button
	divContent.WriteString(`<a class="songAddToPlaylistLink" href="#" data-songid="` + string(song.Id) + `">`)
	divContent.WriteString(`<i class="fas fa-plus"></i>`)
	divContent.WriteString(`</a>`)

	divContent.WriteString(`</div>`)

	divContent.WriteString(`</div>`)

	return divContent.String()
}

func (c *LibraryComponent) refreshPlaylistList() {
	listDiv := c.app.doc.Call("getElementById", "libraryList")

	var divContent strings.Builder

	for _, playlist := range c.app.localDb.OrderedPlaylists {
		if playlist != nil {
			divContent.WriteString(`<div class="item playlistItem">`)

			// Favorite item
			divContent.WriteString(`<div class="itemFavorite"><a class="playlistFavoriteLink" href="#" data-playlistid="` + string(playlist.Id) + `">`)
			if _, ok := c.app.localDb.UserFavoritePlaylistIds[c.app.restClient.UserId()][playlist.Id]; ok {
				divContent.WriteString(`<i class="fas fa-star"></i>`)
			} else {
				divContent.WriteString(`<i class="far fa-star" style="color: #444;"></i>`)
			}
			divContent.WriteString(`</a></div>`)

			// Title item
			divContent.WriteString(`<div class="itemTitle">`)

			divContent.WriteString(`<a class="playlistLink" href="#" data-playlistid="` + string(playlist.Id) + `">` + html.EscapeString(playlist.Name) + `</a><div></div>`)

			divContent.WriteString(`</div>`)

			// Buttons item
			divContent.WriteString(`<div class="itemButtons">`)

			// 'Add to current playlist' button
			divContent.WriteString(`<a class="playlistAddToPlaylistLink" href="#" data-playlistid="` + string(playlist.Id) + `">`)
			divContent.WriteString(`<i class="fas fa-plus"></i>`)
			divContent.WriteString(`</a>`)

			divContent.WriteString(`</div>`)

			divContent.WriteString(`</div>`)
		}
	}
	listDiv.Set("innerHTML", divContent.String())
}

func (c *LibraryComponent) refreshUserList() {
	listDiv := c.app.doc.Call("getElementById", "libraryList")

	var divContent strings.Builder

	for _, user := range c.app.localDb.OrderedUsers {
		if user != nil {
			divContent.WriteString(`<div class="item userItem">`)

			// Title item
			divContent.WriteString(`<div class="itemTitle">`)

			divContent.WriteString(`<a class="userLink" href="#" data-userid="` + string(user.Id) + `">` + html.EscapeString(user.Name) + `</a><div></div>`)

			divContent.WriteString(`</div>`)

			divContent.WriteString(`</div>`)
		}
	}
	listDiv.Set("innerHTML", divContent.String())
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

func (c *LibraryComponent) ShowUsersAction() {
	c.libraryState = libraryState{
		libraryType: libraryTypeUsers,
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
