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
	app           *App
	libraryState  libraryState
	onlyFavorites bool
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
		artistId:    nil,
		albumId:     nil,
		playlistId:  nil,
		userId:      nil,
		nameFilter:  nil,
	}
	c.onlyFavorites = false
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
	listDiv := jst.Document.Call("getElementById", "libraryList")

	var divContent strings.Builder

	var artists []*restApiV1.Artist
	if c.onlyFavorites {
		artists = c.app.localDb.UserOrderedFavoriteArtists[c.app.ConnectedUserId()]
	} else {
		artists = c.app.localDb.OrderedArtists
	}

	for _, artist := range artists {
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
	listDiv.Set("innerHTML", divContent.String())
}

func (c *LibraryComponent) refreshAlbumList() {
	listDiv := jst.Document.Call("getElementById", "libraryList")

	var divContent strings.Builder

	var albums []*restApiV1.Album
	if c.onlyFavorites {
		albums = c.app.localDb.UserOrderedFavoriteAlbums[c.app.ConnectedUserId()]
	} else {
		albums = c.app.localDb.OrderedAlbums
	}

	for _, album := range albums {
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
	listDiv.Set("innerHTML", divContent.String())
}

func (c *LibraryComponent) refreshSongList() {

	listDiv := jst.Document.Call("getElementById", "libraryList")

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
			if c.onlyFavorites {
				songList = c.app.localDb.UserOrderedFavoriteSongs[c.app.ConnectedUserId()]
			} else {
				songList = c.app.localDb.OrderedSongs
			}
		}

		listDiv.Set("innerHTML", "")
		for _, song := range songList {

			// Remove explicit songs if user profile ask for it
			if c.app.HideExplicitSongForConnectedUser() {
				if song.ExplicitFg {
					continue
				}
			}

			listDiv.Call("insertAdjacentHTML", "beforeEnd", c.addSongItem(song))
		}

	} else {
		listDiv.Set("innerHTML", "")
		for _, songId := range c.app.localDb.Playlists[*c.libraryState.playlistId].SongIds {

			// Remove explicit songs if user profile ask for it
			if c.app.HideExplicitSongForConnectedUser() {
				if c.app.localDb.Songs[songId].ExplicitFg {
					continue
				}
			}

			listDiv.Call("insertAdjacentHTML", "beforeEnd", c.addSongItem(c.app.localDb.Songs[songId]))
		}
	}
}

func (c *LibraryComponent) addSongItem(song *restApiV1.Song) string {
	var divContent strings.Builder

	_, favorite := c.app.localDb.UserFavoriteSongIds[c.app.ConnectedUserId()][song.Id]

	songItem := struct {
		SongId    string
		Favorite  bool
		SongName  string
		AlbumId   string
		AlbumName string
		Artists   []struct {
			ArtistId   string
			ArtistName string
		}
	}{
		SongId:   string(song.Id),
		Favorite: favorite,
		SongName: song.Name,
		AlbumId:  string(song.AlbumId),
	}

	if song.AlbumId != restApiV1.UnknownAlbumId {
		songItem.AlbumName = c.app.localDb.Albums[song.AlbumId].Name
	}

	for _, artistId := range song.ArtistIds {
		songItem.Artists = append(songItem.Artists, struct {
			ArtistId   string
			ArtistName string
		}{
			ArtistId:   string(artistId),
			ArtistName: c.app.localDb.Artists[artistId].Name,
		})
	}

	divContent.WriteString(c.app.RenderTemplate(
		&songItem, "songItem.html"),
	)

	// Title item
	separator := ""

	if song.AlbumId != restApiV1.UnknownAlbumId && c.libraryState.albumId == nil {
		divContent.WriteString(`<a class="albumLink" href="#" data-albumid="` + string(song.AlbumId) + `">` + html.EscapeString(c.app.localDb.Albums[song.AlbumId].Name) + `</a>`)
		separator = " / "
	}

	if len(song.ArtistIds) > 0 {
		for _, artistId := range song.ArtistIds {
			if c.libraryState.artistId == nil || (c.libraryState.artistId != nil && artistId != *c.libraryState.artistId) {
				divContent.WriteString(separator + `<a class="artistLink" href="#" data-artistid="` + string(artistId) + `">` + html.EscapeString(c.app.localDb.Artists[artistId].Name) + `</a>`)
				separator = " / "
			}
		}
	}

	divContent.WriteString(`</div></div>`)

	// Buttons item
	divContent.WriteString(`<div class="itemButtons">`)

	// 'Download song' button
	divContent.WriteString(`<a class="songDownloadLink" href="#" data-songid="` + string(song.Id) + `">`)
	divContent.WriteString(`<i class="fas fa-file-download"></i>`)
	divContent.WriteString(`</a>`)

	// 'Add to current playlist' button
	divContent.WriteString(`<a class="songAddToPlaylistLink" href="#" data-songid="` + string(song.Id) + `">`)
	divContent.WriteString(`<i class="fas fa-plus"></i>`)
	divContent.WriteString(`</a>`)

	// 'Play now' button
	divContent.WriteString(`<a class="songPlayNowLink" href="#" data-songid="` + string(song.Id) + `">`)
	divContent.WriteString(`<i class="fas fa-play"></i>`)
	divContent.WriteString(`</a>`)

	divContent.WriteString(`</div>`)

	divContent.WriteString(`</div>`)

	return divContent.String()
}

func (c *LibraryComponent) refreshPlaylistList() {
	listDiv := jst.Document.Call("getElementById", "libraryList")

	var divContent strings.Builder

	var playlists []*restApiV1.Playlist
	if c.onlyFavorites {
		playlists = c.app.localDb.UserOrderedFavoritePlaylists[c.app.ConnectedUserId()]
	} else {
		playlists = c.app.localDb.OrderedPlaylists
	}

	for _, playlist := range playlists {
		if playlist != nil {
			_, favorite := c.app.localDb.UserFavoritePlaylistIds[c.app.ConnectedUserId()][playlist.Id]
			divContent.WriteString(c.app.RenderTemplate(
				struct {
					PlaylistId string
					Favorite   bool
					Name       string
				}{
					PlaylistId: string(playlist.Id),
					Favorite:   favorite,
					Name:       playlist.Name,
				}, "playlistItem.html"),
			)

		}
	}
	listDiv.Set("innerHTML", divContent.String())
}

func (c *LibraryComponent) refreshUserList() {
	listDiv := jst.Document.Call("getElementById", "libraryList")

	var divContent strings.Builder

	for _, user := range c.app.localDb.OrderedUsers {
		if user != nil {
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

func (c *LibraryComponent) FavoritesSwitchAction(this js.Value, args []js.Value) {
	c.onlyFavorites = this.Get("checked").Bool()
	c.RefreshView()
}
