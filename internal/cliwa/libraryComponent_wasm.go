package cliwa

import (
	"fmt"
	"github.com/jypelle/mifasol/internal/cliwa/jst"
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
	libraryType         libraryType
	artistId            *restApiV1.ArtistId
	albumId             *restApiV1.AlbumId
	playlistId          *restApiV1.PlaylistId
	userId              *restApiV1.UserId
	nameFilter          *string
	onlyFavoritesFilter bool
	cachedArtists       []*restApiV1.Artist
	cachedAlbums        []*restApiV1.Album
	cachedSongs         []*restApiV1.Song
	cachedPlaylists     []*restApiV1.Playlist
	cachedUsers         []*restApiV1.User
}

func (c *LibraryComponent) refreshTitle() {

	var title string

	switch c.libraryState.libraryType {
	case libraryTypeArtists:
		if c.libraryState.userId == nil {
			title = `Artists`
		} else {
			title = fmt.Sprintf(`Favorite artists from <span class="userLink">%s</span>`, html.EscapeString(c.app.localDb.Users[*c.libraryState.userId].Name))
		}
	case libraryTypeAlbums:
		if c.libraryState.userId == nil {
			title = `Albums`
		} else {
			title = fmt.Sprintf(`Favorite albums from <span class="userLink">%s</span>`, html.EscapeString(c.app.localDb.Users[*c.libraryState.userId].Name))
		}
	case libraryTypePlaylists:
		if c.libraryState.userId == nil {
			title = `Playlists`
		} else {
			title = fmt.Sprintf(`Favorite playlists from <span class="userLink">%s</span>`, html.EscapeString(c.app.localDb.Users[*c.libraryState.userId].Name))
		}
	case libraryTypeSongs:
		if c.libraryState.userId == nil && c.libraryState.playlistId == nil && c.libraryState.artistId == nil && c.libraryState.albumId == nil {
			title = `Songs`
		}
		if c.libraryState.playlistId != nil {
			title = fmt.Sprintf(`Songs from <span class="playlistLink">%s</span>`, html.EscapeString(c.app.localDb.Playlists[*c.libraryState.playlistId].Name))
		}
		if c.libraryState.userId != nil {
			title = fmt.Sprintf(`Favorite songs from <span class="userLink">%s</span>`, html.EscapeString(c.app.localDb.Users[*c.libraryState.userId].Name))
		}
		if c.libraryState.artistId != nil {
			if *c.libraryState.artistId != restApiV1.UnknownArtistId {
				title = fmt.Sprintf(`Songs from <span class="artistLink">%s</span>`, html.EscapeString(c.app.localDb.Artists[*c.libraryState.artistId].Name))
			} else {
				title = "Songs from unknown artists"
			}
		}
		if c.libraryState.albumId != nil {
			if *c.libraryState.albumId != restApiV1.UnknownAlbumId {
				title = fmt.Sprintf(`Songs from <span class="albumLink">%s</span>`, html.EscapeString(c.app.localDb.Albums[*c.libraryState.albumId].Name))
			} else {
				title = "Songs from unknown album"
			}
		}
	case libraryTypeUsers:
		title = "Users"
	}

	titleSpan := jst.Document.Call("getElementById", "libraryTitle")
	titleSpan.Set("innerHTML", title)
}

type LibraryComponent struct {
	app          *App
	libraryState libraryState
}

func NewLibraryComponent(app *App) *LibraryComponent {
	c := &LibraryComponent{
		app: app,
	}
	c.Reset()

	return c
}

func (c *LibraryComponent) Reset() {
	c.libraryState = libraryState{
		libraryType: libraryTypeArtists,
	}
}

func (c *LibraryComponent) Show() {
	libraryArtistsButton := jst.Document.Call("getElementById", "libraryArtistsButton")
	libraryArtistsButton.Call("addEventListener", "click", c.app.AddEventFunc(c.ShowArtistsAction))
	libraryAlbumsButton := jst.Document.Call("getElementById", "libraryAlbumsButton")
	libraryAlbumsButton.Call("addEventListener", "click", c.app.AddEventFunc(c.ShowAlbumsAction))
	librarySongsButton := jst.Document.Call("getElementById", "librarySongsButton")
	librarySongsButton.Call("addEventListener", "click", c.app.AddEventFunc(c.ShowSongsAction))
	libraryPlaylistsButton := jst.Document.Call("getElementById", "libraryPlaylistsButton")
	libraryPlaylistsButton.Call("addEventListener", "click", c.app.AddEventFunc(c.ShowPlaylistsAction))
	libraryUsersButton := jst.Document.Call("getElementById", "libraryUsersButton")
	libraryUsersButton.Call("addEventListener", "click", c.app.AddEventFunc(c.ShowUsersAction))
	libraryAddToPlaylistButton := jst.Document.Call("getElementById", "libraryAddToPlaylistButton")
	libraryAddToPlaylistButton.Call("addEventListener", "click", c.app.AddEventFunc(c.AddToPlaylistAction))

	libraryFavoritesSwitch := jst.Document.Call("getElementById", "libraryFavoritesSwitch")
	libraryFavoritesSwitch.Call("addEventListener", "click", c.app.AddRichEventFunc(c.FavoritesSwitchAction))

	listDiv := jst.Document.Call("getElementById", "libraryList")
	listDiv.Call("addEventListener", "click", c.app.AddRichEventFunc(func(this js.Value, i []js.Value) {
		link := i[0].Get("target").Call("closest", ".artistLink, .artistAddToPlaylistLink, .albumLink, .albumAddToPlaylistLink, .playlistLink, .playlistFavoriteLink, .playlistAddToPlaylistLink, .playlistLoadToPlaylistLink, .songFavoriteLink, .songAddToPlaylistLink, .songPlayNowLink, .songDownloadLink")
		if !link.Truthy() {
			return
		}
		dataset := link.Get("dataset")

		switch link.Get("className").String() {
		case "artistLink":
			artistId := dataset.Get("artistid").String()
			c.OpenArtistAction(restApiV1.ArtistId(artistId))
		case "artistAddToPlaylistLink":
			artistId := dataset.Get("artistid").String()
			c.app.HomeComponent.CurrentComponent.AddSongsFromArtistAction(restApiV1.ArtistId(artistId))
		case "albumLink":
			albumId := dataset.Get("albumid").String()
			c.OpenAlbumAction(restApiV1.AlbumId(albumId))
		case "albumAddToPlaylistLink":
			albumId := dataset.Get("albumid").String()
			c.app.HomeComponent.CurrentComponent.AddSongsFromAlbumAction(restApiV1.AlbumId(albumId))
		case "playlistLink":
			playlistId := dataset.Get("playlistid").String()
			c.OpenPlaylistAction(restApiV1.PlaylistId(playlistId))
		case "playlistFavoriteLink":
			playlistId := dataset.Get("playlistid").String()
			favoritePlaylistId := restApiV1.FavoritePlaylistId{
				UserId:     c.app.ConnectedUserId(),
				PlaylistId: restApiV1.PlaylistId(playlistId),
			}

			if _, ok := c.app.localDb.UserFavoritePlaylistIds[c.app.ConnectedUserId()][restApiV1.PlaylistId(playlistId)]; ok {
				link.Set("innerHTML", `<i class="far fa-star" style="color: #444;"></i>`)

				_, cliErr := c.app.restClient.DeleteFavoritePlaylist(favoritePlaylistId)
				if cliErr != nil {
					c.app.HomeComponent.MessageComponent.Message("Unable to add playlist to favorites")
					link.Set("innerHTML", `<i class="fas fa-star"></i>`)
					return
				}
				c.app.localDb.RemovePlaylistFromMyFavorite(restApiV1.PlaylistId(playlistId))
			} else {
				link.Set("innerHTML", `<i class="fas fa-star"></i>`)

				_, cliErr := c.app.restClient.CreateFavoritePlaylist(&restApiV1.FavoritePlaylistMeta{Id: favoritePlaylistId})
				if cliErr != nil {
					c.app.HomeComponent.MessageComponent.Message("Unable to remove playlist from favorites")
					link.Set("innerHTML", `<i class="far fa-star" style="color: #444;"></i>`)
					return
				}
				c.app.localDb.AddPlaylistToMyFavorite(restApiV1.PlaylistId(playlistId))

			}

		case "playlistAddToPlaylistLink":
			playlistId := dataset.Get("playlistid").String()
			c.app.HomeComponent.CurrentComponent.AddSongsFromPlaylistAction(restApiV1.PlaylistId(playlistId))
		case "playlistLoadToPlaylistLink":
			playlistId := dataset.Get("playlistid").String()
			c.app.HomeComponent.CurrentComponent.LoadSongsFromPlaylistAction(restApiV1.PlaylistId(playlistId))
		case "songFavoriteLink":
			songId := dataset.Get("songid").String()
			favoriteSongId := restApiV1.FavoriteSongId{
				UserId: c.app.ConnectedUserId(),
				SongId: restApiV1.SongId(songId),
			}

			if _, ok := c.app.localDb.UserFavoriteSongIds[c.app.ConnectedUserId()][restApiV1.SongId(songId)]; ok {
				link.Set("innerHTML", `<i class="far fa-star" style="color: #444;"></i>`)

				_, cliErr := c.app.restClient.DeleteFavoriteSong(favoriteSongId)
				if cliErr != nil {
					c.app.HomeComponent.MessageComponent.Message("Unable to add song to favorites")
					link.Set("innerHTML", `<i class="fas fa-star"></i>`)
					return
				}
				c.app.localDb.RemoveSongFromMyFavorite(restApiV1.SongId(songId))

				logrus.Info("Deactivate")
			} else {
				link.Set("innerHTML", `<i class="fas fa-star"></i>`)

				_, cliErr := c.app.restClient.CreateFavoriteSong(&restApiV1.FavoriteSongMeta{Id: favoriteSongId})
				if cliErr != nil {
					c.app.HomeComponent.MessageComponent.Message("Unable to remove song from favorites")
					link.Set("innerHTML", `<i class="far fa-star" style="color: #444;"></i>`)
					return
				}
				c.app.localDb.AddSongToMyFavorite(restApiV1.SongId(songId))

				logrus.Info("Activate")
			}
		case "songPlayNowLink":
			songId := dataset.Get("songid").String()
			c.app.HomeComponent.PlayerComponent.PlaySongAction(restApiV1.SongId(songId))
		case "songDownloadLink":
			songId := dataset.Get("songid").String()
			song := c.app.localDb.Songs[restApiV1.SongId(songId)]

			token, cliErr := c.app.restClient.GetToken()
			if cliErr != nil {
				return
			}

			anchor := jst.Document.Call("createElement", "a")
			anchor.Set("href", "/api/v1/songContents/"+string(songId)+"?bearer="+token.AccessToken)
			anchor.Set("download", song.Name+song.Format.Extension())
			jst.Document.Get("body").Call("appendChild", anchor)
			anchor.Call("click")
			jst.Document.Get("body").Call("removeChild", anchor)

		case "songAddToPlaylistLink":
			songId := dataset.Get("songid").String()
			c.app.HomeComponent.CurrentComponent.AddSongAction(restApiV1.SongId(songId))
		}
	}))

}

func (c *LibraryComponent) computeCache() {
	// Clear cache
	c.libraryState.cachedArtists = nil
	c.libraryState.cachedAlbums = nil
	c.libraryState.cachedSongs = nil
	c.libraryState.cachedPlaylists = nil
	c.libraryState.cachedUsers = nil

	// Compute cache
	switch c.libraryState.libraryType {
	case libraryTypeArtists:
		c.computeArtistList()
	case libraryTypeAlbums:
		c.computeAlbumList()
	case libraryTypePlaylists:
		c.computePlaylistList()
	case libraryTypeSongs:
		c.computeSongList()
	case libraryTypeUsers:
		c.computeUserList()
	}
}

func (c *LibraryComponent) RefreshView() {

	listDiv := jst.Document.Call("getElementById", "libraryList")
	listDiv.Set("innerHTML", "Loading...")

	c.computeCache()

	var divContent strings.Builder

	// Refresh library list
	switch c.libraryState.libraryType {
	case libraryTypeArtists:

		for _, artist := range c.libraryState.cachedArtists {
			var artistItem struct {
				ArtistId   string
				ArtistName string
			}

			if artist == nil {
				artistItem.ArtistId = string(restApiV1.UnknownArtistId)
				artistItem.ArtistName = "(Unknown artist)"
			} else {
				artistItem.ArtistId = string(artist.Id)
				artistItem.ArtistName = artist.Name
			}

			divContent.WriteString(c.app.RenderTemplate(
				&artistItem, "artistItem.html"),
			)
		}

	case libraryTypeAlbums:

		for _, album := range c.libraryState.cachedAlbums {
			var albumItem struct {
				AlbumId   string
				AlbumName string
				Artists   []struct {
					ArtistId   string
					ArtistName string
				}
			}

			if album == nil {
				albumItem.AlbumId = string(restApiV1.UnknownAlbumId)
				albumItem.AlbumName = "(Unknown album)"
			} else {
				albumItem.AlbumId = string(album.Id)
				albumItem.AlbumName = album.Name
				for _, artistId := range album.ArtistIds {
					albumItem.Artists = append(albumItem.Artists, struct {
						ArtistId   string
						ArtistName string
					}{
						ArtistId:   string(artistId),
						ArtistName: c.app.localDb.Artists[artistId].Name,
					})
				}
			}
			divContent.WriteString(c.app.RenderTemplate(
				&albumItem, "albumItem.html"),
			)
		}
	case libraryTypePlaylists:
		for _, playlist := range c.libraryState.cachedPlaylists {
			if playlist != nil {
				_, favorite := c.app.localDb.UserFavoritePlaylistIds[c.app.ConnectedUserId()][playlist.Id]

				playlistItem := struct {
					PlaylistId string
					Favorite   bool
					Name       string
					OwnerUsers []struct {
						UserId   string
						UserName string
					}
				}{
					PlaylistId: string(playlist.Id),
					Favorite:   favorite,
					Name:       playlist.Name,
				}

				for _, userId := range playlist.OwnerUserIds {
					playlistItem.OwnerUsers = append(playlistItem.OwnerUsers, struct {
						UserId   string
						UserName string
					}{
						UserId:   string(userId),
						UserName: c.app.localDb.Users[userId].Name,
					})
				}

				divContent.WriteString(c.app.RenderTemplate(
					&playlistItem, "playlistItem.html"),
				)

			}
		}
	case libraryTypeSongs:
		for _, song := range c.libraryState.cachedSongs {
			c.addSongItem(&divContent, song)
		}
	case libraryTypeUsers:
		for _, user := range c.libraryState.cachedUsers {
			divContent.WriteString(c.app.RenderTemplate(
				struct {
					UserId string
					Name   string
				}{
					UserId: string(user.Id),
					Name:   user.Name,
				}, "userItem.html"),
			)
		}
	}

	listDiv.Set("innerHTML", divContent.String())

	// Refresh library title
	c.refreshTitle()
}

func (c *LibraryComponent) computeArtistList() {
	if c.libraryState.onlyFavoritesFilter {
		c.libraryState.cachedArtists = c.app.localDb.UserOrderedFavoriteArtists[c.app.ConnectedUserId()]
	} else {
		c.libraryState.cachedArtists = c.app.localDb.OrderedArtists
	}
}

func (c *LibraryComponent) computeAlbumList() {
	if c.libraryState.onlyFavoritesFilter {
		c.libraryState.cachedAlbums = c.app.localDb.UserOrderedFavoriteAlbums[c.app.ConnectedUserId()]
	} else {
		c.libraryState.cachedAlbums = c.app.localDb.OrderedAlbums
	}
}

func (c *LibraryComponent) computeSongList() {

	var songList []*restApiV1.Song

	c.libraryState.cachedSongs = nil

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
			if c.libraryState.onlyFavoritesFilter {
				songList = c.app.localDb.UserOrderedFavoriteSongs[c.app.ConnectedUserId()]
			} else {
				songList = c.app.localDb.OrderedSongs
			}
		}

		for _, song := range songList {

			// Remove explicit songs if user profile ask for it
			if c.app.HideExplicitSongForConnectedUser() {
				if song.ExplicitFg {
					continue
				}
			}

			// Remove non favorite songs if user ask for it
			if c.libraryState.onlyFavoritesFilter {
				_, favorite := c.app.localDb.UserFavoriteSongIds[c.app.ConnectedUserId()][song.Id]
				if !favorite {
					continue
				}
			}

			c.libraryState.cachedSongs = append(c.libraryState.cachedSongs, song)
		}

	} else {

		for _, songId := range c.app.localDb.Playlists[*c.libraryState.playlistId].SongIds {

			// Remove explicit songs if user profile ask for it
			if c.app.HideExplicitSongForConnectedUser() {
				if c.app.localDb.Songs[songId].ExplicitFg {
					continue
				}
			}

			// Remove non favorite songs if user ask for it
			if c.libraryState.onlyFavoritesFilter {
				_, favorite := c.app.localDb.UserFavoriteSongIds[c.app.ConnectedUserId()][songId]
				if !favorite {
					continue
				}
			}
			c.libraryState.cachedSongs = append(c.libraryState.cachedSongs, c.app.localDb.Songs[songId])

		}
	}
}

func (c *LibraryComponent) addSongItem(divContent *strings.Builder, song *restApiV1.Song) {

	_, favorite := c.app.localDb.UserFavoriteSongIds[c.app.ConnectedUserId()][song.Id]

	songItem := struct {
		SongId    string
		Favorite  bool
		SongName  string
		AlbumId   *string
		AlbumName string
		Artists   []struct {
			ArtistId   string
			ArtistName string
		}
	}{
		SongId:   string(song.Id),
		Favorite: favorite,
		SongName: song.Name,
	}

	if song.AlbumId != restApiV1.UnknownAlbumId && c.libraryState.albumId == nil {
		songItem.AlbumName = c.app.localDb.Albums[song.AlbumId].Name
		songItem.AlbumId = (*string)(&song.AlbumId)
	}

	for _, artistId := range song.ArtistIds {
		if c.libraryState.artistId == nil || (c.libraryState.artistId != nil && artistId != *c.libraryState.artistId) {
			songItem.Artists = append(songItem.Artists, struct {
				ArtistId   string
				ArtistName string
			}{
				ArtistId:   string(artistId),
				ArtistName: c.app.localDb.Artists[artistId].Name,
			})
		}
	}

	divContent.WriteString(c.app.RenderTemplate(
		&songItem, "songItem.html"),
	)
}

func (c *LibraryComponent) computePlaylistList() {
	if c.libraryState.onlyFavoritesFilter {
		c.libraryState.cachedPlaylists = c.app.localDb.UserOrderedFavoritePlaylists[c.app.ConnectedUserId()]
	} else {
		c.libraryState.cachedPlaylists = c.app.localDb.OrderedPlaylists
	}
}

func (c *LibraryComponent) computeUserList() {
	c.libraryState.cachedUsers = c.app.localDb.OrderedUsers
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

func (c *LibraryComponent) AddToPlaylistAction() {
	if len(c.libraryState.cachedSongs) > 0 {
		c.app.HomeComponent.CurrentComponent.AddSongsAction(c.libraryState.cachedSongs)
	}
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

func (c *LibraryComponent) FavoritesSwitchAction(this js.Value, args []js.Value) {
	c.libraryState.onlyFavoritesFilter = this.Get("checked").Bool()
	c.RefreshView()
}
