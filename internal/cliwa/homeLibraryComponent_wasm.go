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
	LibraryTypeArtists libraryType = iota
	LibraryTypeAlbums
	LibraryTypePlaylists
	LibraryTypeSongs
	LibraryTypeUsers
)

const LibraryPageSize = 50

type libraryState struct {
	libraryType         libraryType
	artistId            *restApiV1.ArtistId
	albumId             *restApiV1.AlbumId
	playlistId          *restApiV1.PlaylistId
	userId              *restApiV1.UserId
	nameFilter          *string
	onlyFavoritesFilter bool
	displayedPage       int
	cachedArtists       []*restApiV1.Artist
	cachedAlbums        []*restApiV1.Album
	cachedSongs         []*restApiV1.Song
	cachedPlaylists     []*restApiV1.Playlist
	cachedUsers         []*restApiV1.User
}

func (s *libraryState) cachedSize() int {
	if len(s.cachedArtists) > 0 {
		return len(s.cachedArtists)
	}
	if len(s.cachedAlbums) > 0 {
		return len(s.cachedAlbums)
	}
	if len(s.cachedSongs) > 0 {
		return len(s.cachedSongs)
	}
	if len(s.cachedPlaylists) > 0 {
		return len(s.cachedPlaylists)
	}
	if len(s.cachedUsers) > 0 {
		return len(s.cachedUsers)
	}
	return 0
}

type LibraryComponent struct {
	app          *App
	libraryState libraryState
}

func NewHomeLibraryComponent(app *App) *LibraryComponent {
	c := &LibraryComponent{
		app: app,
		libraryState: libraryState{
			libraryType: LibraryTypeArtists,
		},
	}

	return c
}

func (c *LibraryComponent) Show() {
	div := jst.Document.Call("getElementById", "libraryComponent")
	div.Set("innerHTML", c.app.RenderTemplate(
		nil, "home/library/index"),
	)

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

	librarySearchInput := jst.Document.Call("getElementById", "librarySearchInput")
	librarySearchInput.Call("addEventListener", "input", c.app.AddEventFunc(c.SearchAction))
	libraryOnlyFavoritesButton := jst.Document.Call("getElementById", "libraryOnlyFavoritesButton")
	libraryOnlyFavoritesButton.Call("addEventListener", "click", c.app.AddEventFunc(c.FavoritesSwitchAction))

	libraryList := jst.Document.Call("getElementById", "libraryList")
	libraryList.Call("addEventListener", "scroll", c.app.AddRichEventFunc(func(this js.Value, i []js.Value) {
		scrollHeight := libraryList.Get("scrollHeight").Int()
		scrollTop := libraryList.Get("scrollTop").Int()
		clientHeight := libraryList.Get("clientHeight").Int()
		if scrollTop+clientHeight >= scrollHeight-5 {
			if LibraryPageSize*(c.libraryState.displayedPage+2) <= c.libraryState.cachedSize() {
				c.libraryState.displayedPage++
				logrus.Infof("scroll: Down %d / %d / %d / %d", c.libraryState.displayedPage, scrollHeight, scrollTop+clientHeight, clientHeight)
				c.updateLibraryList(1)
			}
		} else if scrollTop == 0 {
			if c.libraryState.displayedPage > 0 {
				c.libraryState.displayedPage--
				logrus.Infof("scroll: Up %d / %d / %d / %d", c.libraryState.displayedPage, scrollHeight, scrollTop+clientHeight, clientHeight)
				c.updateLibraryList(-1)
			}
		}
	}))
	libraryList.Call("addEventListener", "click", c.app.AddRichEventFunc(func(this js.Value, i []js.Value) {
		link := i[0].Get("target").Call("closest",
			".artistLink, .artistEditLink, .artistDeleteLink, .artistAddToPlaylistLink, "+
				".albumLink, .albumEditLink, .albumDeleteLink, .albumAddToPlaylistLink, "+
				".playlistLink, .playlistEditLink, .playlistDeleteLink, .playlistFavoriteLink, .playlistAddToPlaylistLink, .playlistLoadToPlaylistLink, "+
				".songEditLink, .songDeleteLink, .songFavoriteLink, .songAddToPlaylistLink, .songPlayNowLink, .songDownloadLink, "+
				".userEditLink, .userDeleteLink")
		if !link.Truthy() {
			return
		}
		dataset := link.Get("dataset")

		switch link.Get("className").String() {
		case "artistLink":
			artistId := dataset.Get("artistid").String()
			c.OpenArtistAction(restApiV1.ArtistId(artistId))
		case "artistEditLink":
			artistId := restApiV1.ArtistId(dataset.Get("artistid").String())
			if artistId != restApiV1.UnknownArtistId {
				component := NewHomeArtistEditComponent(c.app, artistId, c.app.localDb.Artists[artistId].ArtistMeta)

				c.app.HomeComponent.OpenModal()
				component.Show()
			}
		case "artistDeleteLink":
			artistId := restApiV1.ArtistId(dataset.Get("artistid").String())
			component := NewHomeConfirmDeleteComponent(c.app, artistId)
			c.app.HomeComponent.OpenModal()
			component.Show()

		case "artistAddToPlaylistLink":
			artistId := dataset.Get("artistid").String()
			c.app.HomeComponent.CurrentComponent.AddSongsFromArtistAction(restApiV1.ArtistId(artistId))
		case "albumLink":
			albumId := dataset.Get("albumid").String()
			c.OpenAlbumAction(restApiV1.AlbumId(albumId))
		case "albumEditLink":
			albumId := restApiV1.AlbumId(dataset.Get("albumid").String())
			if albumId != restApiV1.UnknownAlbumId {
				component := NewHomeAlbumEditComponent(c.app, albumId, c.app.localDb.Albums[albumId].AlbumMeta)

				c.app.HomeComponent.OpenModal()
				component.Show()
			}
		case "albumDeleteLink":
			albumId := restApiV1.AlbumId(dataset.Get("albumid").String())
			component := NewHomeConfirmDeleteComponent(c.app, albumId)

			c.app.HomeComponent.OpenModal()
			component.Show()

		case "albumAddToPlaylistLink":
			albumId := restApiV1.AlbumId(dataset.Get("albumid").String())
			c.app.HomeComponent.CurrentComponent.AddSongsFromAlbumAction(albumId)
		case "playlistLink":
			playlistId := restApiV1.PlaylistId(dataset.Get("playlistid").String())
			c.OpenPlaylistAction(playlistId)
		case "playlistEditLink":
			playlistId := restApiV1.PlaylistId(dataset.Get("playlistid").String())
			component := NewHomePlaylistEditComponent(c.app, playlistId, c.app.localDb.Playlists[playlistId].PlaylistMeta)
			c.app.HomeComponent.OpenModal()
			component.Show()

		case "playlistDeleteLink":
			playlistId := restApiV1.PlaylistId(dataset.Get("playlistid").String())
			component := NewHomeConfirmDeleteComponent(c.app, playlistId)

			c.app.HomeComponent.OpenModal()
			component.Show()
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
		case "songEditLink":
			songId := restApiV1.SongId(dataset.Get("songid").String())
			component := NewHomeSongEditComponent(c.app, songId, c.app.localDb.Songs[songId].SongMeta)
			c.app.HomeComponent.OpenModal()
			component.Show()
		case "songDeleteLink":
			songId := restApiV1.SongId(dataset.Get("songid").String())
			component := NewHomeConfirmDeleteComponent(c.app, songId)

			c.app.HomeComponent.OpenModal()
			component.Show()
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
		case "userEditLink":
			//userId := dataset.Get("userid").String()
		case "userDeleteLink":
			userId := restApiV1.UserId(dataset.Get("userid").String())
			component := NewHomeConfirmDeleteComponent(c.app, userId)

			c.app.HomeComponent.OpenModal()
			component.Show()
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
	case LibraryTypeArtists:
		c.computeArtistList()
	case LibraryTypeAlbums:
		c.computeAlbumList()
	case LibraryTypePlaylists:
		c.computePlaylistList()
	case LibraryTypeSongs:
		c.computeSongList()
	case LibraryTypeUsers:
		c.computeUserList()
	}
}

func (c *LibraryComponent) RefreshView() {
	libraryList := jst.Document.Call("getElementById", "libraryList")
	libraryList.Set("innerHTML", "Loading...")
	c.libraryState.displayedPage = 0

	c.computeCache()

	// Update library title
	c.updateTitle()

	// Update buttons
	libraryOnlyFavoritesButton := jst.Document.Call("getElementById", "libraryOnlyFavoritesButton")
	if c.libraryState.onlyFavoritesFilter {
		libraryOnlyFavoritesButton.Set("innerHTML", `<i class="fas fa-star-half-alt"></i>`)
	} else {
		libraryOnlyFavoritesButton.Set("innerHTML", `<i class="fas fa-star"></i>`)
	}
	libraryAddToPlaylistButton := jst.Document.Call("getElementById", "libraryAddToPlaylistButton")
	if len(c.libraryState.cachedSongs) > 0 {
		libraryAddToPlaylistButton.Set("disabled", false)
	} else {
		libraryAddToPlaylistButton.Set("disabled", true)
	}

	// Update list
	c.updateLibraryList(0)
}

func (c *LibraryComponent) computeArtistList() {
	var artistList []*restApiV1.Artist

	if c.libraryState.onlyFavoritesFilter {
		artistList = c.app.localDb.UserOrderedFavoriteArtists[c.app.ConnectedUserId()]
	} else {
		artistList = c.app.localDb.OrderedArtists
	}

	if c.libraryState.nameFilter != nil {
		lowerNameFilter := strings.ToLower(*c.libraryState.nameFilter)
		for _, artist := range artistList {
			if artist != nil && !strings.Contains(strings.ToLower(artist.Name), lowerNameFilter) {
				continue
			}

			c.libraryState.cachedArtists = append(c.libraryState.cachedArtists, artist)
		}
	} else {
		c.libraryState.cachedArtists = artistList
	}
}

func (c *LibraryComponent) computeAlbumList() {
	var albumList []*restApiV1.Album

	if c.libraryState.onlyFavoritesFilter {
		albumList = c.app.localDb.UserOrderedFavoriteAlbums[c.app.ConnectedUserId()]
	} else {
		albumList = c.app.localDb.OrderedAlbums
	}

	if c.libraryState.nameFilter != nil {
		lowerNameFilter := strings.ToLower(*c.libraryState.nameFilter)
		for _, album := range albumList {
			if album != nil && !strings.Contains(strings.ToLower(album.Name), lowerNameFilter) {
				continue
			}

			c.libraryState.cachedAlbums = append(c.libraryState.cachedAlbums, album)
		}
	} else {
		c.libraryState.cachedAlbums = albumList
	}
}

func (c *LibraryComponent) computeSongList() {

	var songList []*restApiV1.Song

	var lowerNameFilter string
	if c.libraryState.nameFilter != nil {
		lowerNameFilter = strings.ToLower(*c.libraryState.nameFilter)
	}

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

			// Remove non matching song name
			if lowerNameFilter != "" && !strings.Contains(strings.ToLower(song.Name), lowerNameFilter) {
				continue
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

			// Remove non matching song name
			if lowerNameFilter != "" && !strings.Contains(strings.ToLower(c.app.localDb.Songs[songId].Name), lowerNameFilter) {
				continue
			}

			c.libraryState.cachedSongs = append(c.libraryState.cachedSongs, c.app.localDb.Songs[songId])

		}
	}
}

func (c *LibraryComponent) computePlaylistList() {
	var playlistList []*restApiV1.Playlist

	if c.libraryState.onlyFavoritesFilter {
		playlistList = c.app.localDb.UserOrderedFavoritePlaylists[c.app.ConnectedUserId()]
	} else {
		playlistList = c.app.localDb.OrderedPlaylists
	}

	if c.libraryState.nameFilter != nil {
		lowerNameFilter := strings.ToLower(*c.libraryState.nameFilter)
		for _, playlist := range playlistList {
			if !strings.Contains(strings.ToLower(playlist.Name), lowerNameFilter) {
				continue
			}

			c.libraryState.cachedPlaylists = append(c.libraryState.cachedPlaylists, playlist)
		}
	} else {
		c.libraryState.cachedPlaylists = playlistList
	}
}

func (c *LibraryComponent) computeUserList() {
	var userList []*restApiV1.User

	userList = c.app.localDb.OrderedUsers

	if c.libraryState.nameFilter != nil {
		lowerNameFilter := strings.ToLower(*c.libraryState.nameFilter)
		for _, user := range userList {
			if !strings.Contains(strings.ToLower(user.Name), lowerNameFilter) {
				continue
			}

			c.libraryState.cachedUsers = append(c.libraryState.cachedUsers, user)
		}
	} else {
		c.libraryState.cachedUsers = userList
	}
}

func (c *LibraryComponent) updateTitle() {

	var title string

	switch c.libraryState.libraryType {
	case LibraryTypeArtists:
		if c.libraryState.userId == nil {
			title = `Artists`
		} else {
			title = fmt.Sprintf(`Favorite artists from <span class="userLink">%s</span>`, html.EscapeString(c.app.localDb.Users[*c.libraryState.userId].Name))
		}
	case LibraryTypeAlbums:
		if c.libraryState.userId == nil {
			title = `Albums`
		} else {
			title = fmt.Sprintf(`Favorite albums from <span class="userLink">%s</span>`, html.EscapeString(c.app.localDb.Users[*c.libraryState.userId].Name))
		}
	case LibraryTypePlaylists:
		if c.libraryState.userId == nil {
			title = `Playlists`
		} else {
			title = fmt.Sprintf(`Favorite playlists from <span class="userLink">%s</span>`, html.EscapeString(c.app.localDb.Users[*c.libraryState.userId].Name))
		}
	case LibraryTypeSongs:
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
	case LibraryTypeUsers:
		title = "Users"
	}

	titleSpan := jst.Document.Call("getElementById", "libraryTitle")
	titleSpan.Set("innerHTML", title)
}

func (c *LibraryComponent) updateLibraryList(direction int) {
	libraryList := jst.Document.Call("getElementById", "libraryList")
	if direction == 0 {
		libraryList.Set("scrollTop", 0)
	}

	var divContentPreviousPage strings.Builder
	var divContentCurrentPage strings.Builder
	var divContentNextPage strings.Builder

	// Refresh library list
	minIdx := LibraryPageSize * (c.libraryState.displayedPage - 1)
	if minIdx < 0 {
		minIdx = 0
	}
	maxIdx := LibraryPageSize * (c.libraryState.displayedPage + 2)
	if maxIdx > c.libraryState.cachedSize() {
		maxIdx = c.libraryState.cachedSize()
	}

	step1Idx := LibraryPageSize*c.libraryState.displayedPage - minIdx
	step2Idx := LibraryPageSize*(c.libraryState.displayedPage+1) - minIdx

	switch c.libraryState.libraryType {
	case LibraryTypeArtists:
		for idx, artist := range c.libraryState.cachedArtists[minIdx:maxIdx] {
			if idx < step1Idx {
				c.addArtistItem(&divContentPreviousPage, artist)
			} else if idx < step2Idx {
				c.addArtistItem(&divContentCurrentPage, artist)
			} else {
				c.addArtistItem(&divContentNextPage, artist)
			}
		}
	case LibraryTypeAlbums:
		for idx, album := range c.libraryState.cachedAlbums[minIdx:maxIdx] {
			if idx < step1Idx {
				c.addAlbumItem(&divContentPreviousPage, album)
			} else if idx < step2Idx {
				c.addAlbumItem(&divContentCurrentPage, album)
			} else {
				c.addAlbumItem(&divContentNextPage, album)
			}
		}
	case LibraryTypePlaylists:
		for idx, playlist := range c.libraryState.cachedPlaylists[minIdx:maxIdx] {
			if idx < step1Idx {
				c.addPlaylistItem(&divContentPreviousPage, playlist)
			} else if idx < step2Idx {
				c.addPlaylistItem(&divContentCurrentPage, playlist)
			} else {
				c.addPlaylistItem(&divContentNextPage, playlist)
			}
		}
	case LibraryTypeSongs:
		for idx, song := range c.libraryState.cachedSongs[minIdx:maxIdx] {
			if idx < step1Idx {
				c.addSongItem(&divContentPreviousPage, song)
			} else if idx < step2Idx {
				c.addSongItem(&divContentCurrentPage, song)
			} else {
				c.addSongItem(&divContentNextPage, song)
			}
		}
	case LibraryTypeUsers:
		for idx, user := range c.libraryState.cachedUsers {
			if idx < step1Idx {
				c.addUserItem(&divContentPreviousPage, user)
			} else if idx < step2Idx {
				c.addUserItem(&divContentCurrentPage, user)
			} else {
				c.addUserItem(&divContentNextPage, user)
			}
		}
	}
	var newScrollTop int
	libraryList.Set("innerHTML", divContentPreviousPage.String())
	if direction == -1 {
		newScrollTop = libraryList.Get("scrollHeight").Int()
	}
	libraryList.Call("insertAdjacentHTML", "beforeEnd", divContentCurrentPage.String())
	if direction == 1 {
		newScrollTop = libraryList.Get("scrollHeight").Int() - libraryList.Get("clientHeight").Int()
	}
	libraryList.Call("insertAdjacentHTML", "beforeEnd", divContentNextPage.String())

	if direction != 0 {
		libraryList.Set("scrollTop", newScrollTop)
		//		logrus.Infof("Set scroll bottom: %d vs currentScrollHeight: %d", newScrollTop + libraryList.Get("clientHeight").Int(), libraryList.Get("scrollHeight").Int())
	}
}

func (c *LibraryComponent) addArtistItem(divContent *strings.Builder, artist *restApiV1.Artist) {
	var artistItem struct {
		ArtistId   string
		ArtistName string
		IsEditable bool
	}

	if artist == nil {
		artistItem.ArtistId = string(restApiV1.UnknownArtistId)
		artistItem.ArtistName = "(Unknown artist)"
		artistItem.IsEditable = false
	} else {
		artistItem.ArtistId = string(artist.Id)
		artistItem.ArtistName = artist.Name
		artistItem.IsEditable = c.app.IsConnectedUserAdmin()
	}

	divContent.WriteString(c.app.RenderTemplate(
		&artistItem, "home/library/artistItem"),
	)
}

func (c *LibraryComponent) addAlbumItem(divContent *strings.Builder, album *restApiV1.Album) {
	var albumItem struct {
		AlbumId   string
		AlbumName string
		Artists   []struct {
			ArtistId   string
			ArtistName string
		}
		IsEditable bool
	}

	if album == nil {
		albumItem.AlbumId = string(restApiV1.UnknownAlbumId)
		albumItem.AlbumName = "(Unknown album)"
		albumItem.IsEditable = false
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
		albumItem.IsEditable = c.app.IsConnectedUserAdmin()
	}
	divContent.WriteString(c.app.RenderTemplate(&albumItem, "home/library/albumItem"))

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
		IsEditable bool
	}{
		SongId:     string(song.Id),
		Favorite:   favorite,
		SongName:   song.Name,
		IsEditable: c.app.IsConnectedUserAdmin(),
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
		&songItem, "home/library/songItem"),
	)
}

func (c *LibraryComponent) addPlaylistItem(divContent *strings.Builder, playlist *restApiV1.Playlist) {
	_, favorite := c.app.localDb.UserFavoritePlaylistIds[c.app.ConnectedUserId()][playlist.Id]

	playlistItem := struct {
		PlaylistId string
		Favorite   bool
		Name       string
		OwnerUsers []struct {
			UserId   string
			UserName string
		}
		IsEditable  bool
		IsDeletable bool
	}{
		PlaylistId:  string(playlist.Id),
		Favorite:    favorite,
		Name:        playlist.Name,
		IsEditable:  c.app.IsConnectedUserAdmin() || c.app.localDb.IsPlaylistOwnedBy(playlist.Id, c.app.ConnectedUserId()),
		IsDeletable: playlist.Id != restApiV1.IncomingPlaylistId && (c.app.IsConnectedUserAdmin() || c.app.localDb.IsPlaylistOwnedBy(playlist.Id, c.app.ConnectedUserId())),
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
		&playlistItem, "home/library/playlistItem"),
	)
}

func (c *LibraryComponent) addUserItem(divContent *strings.Builder, user *restApiV1.User) {
	divContent.WriteString(c.app.RenderTemplate(
		struct {
			UserId     string
			Name       string
			IsEditable bool
		}{
			UserId:     string(user.Id),
			Name:       user.Name,
			IsEditable: c.app.IsConnectedUserAdmin(),
		}, "home/library/userItem"),
	)
}

func (c *LibraryComponent) ShowArtistsAction() {
	c.libraryState = libraryState{
		libraryType: LibraryTypeArtists,
	}
	jst.Document.Call("getElementById", "librarySearchInput").Set("value", "")
	c.RefreshView()
}

func (c *LibraryComponent) ShowAlbumsAction() {
	c.libraryState = libraryState{
		libraryType: LibraryTypeAlbums,
	}
	jst.Document.Call("getElementById", "librarySearchInput").Set("value", "")
	c.RefreshView()
}

func (c *LibraryComponent) ShowSongsAction() {
	c.libraryState = libraryState{
		libraryType: LibraryTypeSongs,
	}
	jst.Document.Call("getElementById", "librarySearchInput").Set("value", "")
	c.RefreshView()
}

func (c *LibraryComponent) ShowPlaylistsAction() {
	c.libraryState = libraryState{
		libraryType: LibraryTypePlaylists,
	}
	jst.Document.Call("getElementById", "librarySearchInput").Set("value", "")
	c.RefreshView()
}

func (c *LibraryComponent) ShowUsersAction() {
	c.libraryState = libraryState{
		libraryType: LibraryTypeUsers,
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
		libraryType: LibraryTypeSongs,
		albumId:     &albumId,
	}
	jst.Document.Call("getElementById", "librarySearchInput").Set("value", "")
	c.RefreshView()
}

func (c *LibraryComponent) OpenArtistAction(artistId restApiV1.ArtistId) {
	c.libraryState = libraryState{
		libraryType: LibraryTypeSongs,
		artistId:    &artistId,
	}
	jst.Document.Call("getElementById", "librarySearchInput").Set("value", "")
	c.RefreshView()
}

func (c *LibraryComponent) OpenPlaylistAction(playlistId restApiV1.PlaylistId) {
	c.libraryState = libraryState{
		libraryType: LibraryTypeSongs,
		playlistId:  &playlistId,
	}
	jst.Document.Call("getElementById", "librarySearchInput").Set("value", "")
	c.RefreshView()
}

func (c *LibraryComponent) FavoritesSwitchAction() {
	c.libraryState.onlyFavoritesFilter = !c.libraryState.onlyFavoritesFilter
	c.RefreshView()
}

func (c *LibraryComponent) SearchAction() {
	librarySearchInput := jst.Document.Call("getElementById", "librarySearchInput")
	nameFilter := librarySearchInput.Get("value").String()

	if nameFilter != "" {
		c.libraryState.nameFilter = &nameFilter
	} else {
		c.libraryState.nameFilter = nil
	}
	c.RefreshView()
}
