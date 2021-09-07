package cliwa

import (
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/sirupsen/logrus"
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

	listDiv := c.doc.Call("getElementById", "libraryList")

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
			c.playSong(restApiV1.SongId(songId))
		case "artistLink":
			artistId := dataset.Get("artistid").String()
			c.openArtist(restApiV1.ArtistId(artistId))
		case "albumLink":
			albumId := dataset.Get("albumid").String()
			c.openAlbum(restApiV1.AlbumId(albumId))
		case "playlistLink":
			playlistId := dataset.Get("playlistid").String()
			c.openPlaylist(restApiV1.PlaylistId(playlistId))
		case "songFavoriteLink":
			songId := dataset.Get("songid").String()
			favoriteSongId := restApiV1.FavoriteSongId{
				UserId: c.restClient.UserId(),
				SongId: restApiV1.SongId(songId),
			}

			go func() {

				if _, ok := c.localDb.UserFavoriteSongIds[c.restClient.UserId()][restApiV1.SongId(songId)]; ok {
					link.Set("innerHTML", `<i class="far fa-star" style="color: #444;"></i>`)

					_, cliErr := c.restClient.DeleteFavoriteSong(favoriteSongId)
					if cliErr != nil {
						c.Message("Unable to add song to favorites")
						link.Set("innerHTML", `<i class="fas fa-star"></i>`)
						return
					}
					c.localDb.RemoveSongFromMyFavorite(restApiV1.SongId(songId))

					logrus.Info("Deactivate")
				} else {
					link.Set("innerHTML", `<i class="fas fa-star"></i>`)

					_, cliErr := c.restClient.CreateFavoriteSong(&restApiV1.FavoriteSongMeta{Id: favoriteSongId})
					if cliErr != nil {
						c.Message("Unable to remove song from favorites")
						link.Set("innerHTML", `<i class="far fa-star" style="color: #444;"></i>`)
						return
					}
					c.localDb.AddSongToMyFavorite(restApiV1.SongId(songId))

					logrus.Info("Activate")
				}
			}()

			/*

							song := c.songs[c.list.GetCurrentItem()]
				if song != nil {
					myFavoriteSongIds := c.uiApp.LocalDb().UserFavoriteSongIds[c.uiApp.ConnectedUserId()]
					favoriteSongId := restApiV1.FavoriteSongId{
						UserId: c.uiApp.ConnectedUserId(),
						SongId: song.Id,
					}
					if _, ok := myFavoriteSongIds[song.Id]; ok {
						_, cliErr := c.uiApp.restClient.DeleteFavoriteSong(favoriteSongId)
						if cliErr != nil {
							c.uiApp.ClientErrorMessage("Unable to add song to favorites", cliErr)
						}
						c.uiApp.LocalDb().RemoveSongFromMyFavorite(song.Id)
						c.RefreshList()
					} else {
						_, cliErr := c.uiApp.restClient.CreateFavoriteSong(&restApiV1.FavoriteSongMeta{Id: favoriteSongId})
						if cliErr != nil {
							c.uiApp.ClientErrorMessage("Unable to remove song from favorites", cliErr)
						}
						c.uiApp.LocalDb().AddSongToMyFavorite(song.Id)
						c.RefreshList()
					}
					if !(currentFilter.userId != nil && *currentFilter.userId == c.uiApp.ConnectedUserId()) {
						c.list.SetCurrentItem(c.list.GetCurrentItem() + 1)
					}
				}

			*/

		}

		return nil
	}))

	go func() {
		c.Reload()
	}()

}
