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
	div := jst.Id("libraryComponent")
	div.Set("innerHTML", c.app.RenderTemplate(
		nil, "home/library/index"),
	)

	libraryArtistsButton := jst.Id("libraryArtistsButton")
	libraryArtistsButton.Call("addEventListener", "click", c.app.AddEventFunc(c.ShowArtistsAction))
	libraryAlbumsButton := jst.Id("libraryAlbumsButton")
	libraryAlbumsButton.Call("addEventListener", "click", c.app.AddEventFunc(c.ShowAlbumsAction))
	librarySongsButton := jst.Id("librarySongsButton")
	librarySongsButton.Call("addEventListener", "click", c.app.AddEventFunc(c.ShowSongsAction))
	libraryPlaylistsButton := jst.Id("libraryPlaylistsButton")
	libraryPlaylistsButton.Call("addEventListener", "click", c.app.AddEventFunc(c.ShowPlaylistsAction))
	libraryUsersButton := jst.Id("libraryUsersButton")
	libraryUsersButton.Call("addEventListener", "click", c.app.AddEventFunc(c.ShowUsersAction))
	libraryAddToPlaylistButton := jst.Id("libraryAddToPlaylistButton")
	libraryAddToPlaylistButton.Call("addEventListener", "click", c.app.AddEventFunc(c.AddToPlaylistAction))
	libraryCreateButton := jst.Id("libraryCreateButton")
	libraryCreateButton.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		if c.libraryState.libraryType == LibraryTypeUsers {
			component := NewHomeUserCreateComponent(c.app)
			c.app.HomeComponent.OpenModal()
			component.Show()
		}
	}))

	librarySearchInput := jst.Id("librarySearchInput")
	librarySearchInput.Call("addEventListener", "input", c.app.AddEventFunc(c.SearchAction))
	libraryOnlyFavoritesButton := jst.Id("libraryOnlyFavoritesButton")
	libraryOnlyFavoritesButton.Call("addEventListener", "click", c.app.AddEventFunc(c.FavoritesSwitchAction))

	libraryList := jst.Id("libraryList")
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
				component := NewHomeArtistEditComponent(c.app, artistId, &c.app.localDb.Artists[artistId].ArtistMeta)

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
				component := NewHomeAlbumEditComponent(c.app, albumId, &c.app.localDb.Albums[albumId].AlbumMeta)

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
			component := NewHomePlaylistEditComponent(c.app, playlistId, &c.app.localDb.Playlists[playlistId].PlaylistMeta)
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
			playlistId := restApiV1.PlaylistId(dataset.Get("playlistid").String())
			c.app.HomeComponent.CurrentComponent.AddSongsFromPlaylistAction(playlistId)
		case "playlistLoadToPlaylistLink":
			playlistId := restApiV1.PlaylistId(dataset.Get("playlistid").String())
			c.app.HomeComponent.CurrentComponent.LoadSongsFromPlaylistAction(playlistId)
		case "songEditLink":
			songId := restApiV1.SongId(dataset.Get("songid").String())
			component := NewHomeSongEditComponent(c.app, songId, &c.app.localDb.Songs[songId].SongMeta)
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
			songId := restApiV1.SongId(dataset.Get("songid").String())
			c.app.HomeComponent.PlayerComponent.PlaySongAction(songId)
		case "songDownloadLink":
			songId := restApiV1.SongId(dataset.Get("songid").String())
			song := c.app.localDb.Songs[songId]

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
			songId := restApiV1.SongId(dataset.Get("songid").String())
			c.app.HomeComponent.CurrentComponent.AddSongAction(songId)
		case "userEditLink":
			userId := restApiV1.UserId(dataset.Get("userid").String())
			component := NewHomeUserEditComponent(c.app, userId, &c.app.localDb.Users[userId].UserMeta)
			c.app.HomeComponent.OpenModal()
			component.Show()

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
	libraryList := jst.Id("libraryList")
	libraryList.Set("innerHTML", "Loading...")
	c.libraryState.displayedPage = 0

	c.computeCache()

	// Update library title
	c.updateTitle()

	// Update buttons
	libraryOnlyFavoritesButton := jst.Id("libraryOnlyFavoritesButton")
	if c.libraryState.onlyFavoritesFilter {
		libraryOnlyFavoritesButton.Set("innerHTML", `<i class="fas fa-star-half-alt"></i>`)
	} else {
		libraryOnlyFavoritesButton.Set("innerHTML", `<i class="fas fa-star"></i>`)
	}
	libraryAddToPlaylistButton := jst.Id("libraryAddToPlaylistButton")
	if len(c.libraryState.cachedSongs) > 0 {
		libraryAddToPlaylistButton.Set("disabled", false)
	} else {
		libraryAddToPlaylistButton.Set("disabled", true)
	}
	libraryCreateButton := jst.Id("libraryCreateButton")
	if c.libraryState.libraryType == LibraryTypeUsers {
		libraryCreateButton.Set("disabled", false)
	} else {
		libraryCreateButton.Set("disabled", true)
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
		lowerNameFilter := strings.TrimSpace(strings.ToLower(*c.libraryState.nameFilter))
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

	titleSpan := jst.Id("libraryTitle")
	titleSpan.Set("innerHTML", title)
}

func (c *LibraryComponent) updateLibraryList(direction int) {
	libraryList := jst.Id("libraryList")
	if direction == 0 {
		libraryList.Set("scrollTop", 0)
	}

	var divContentPreviousPage string
	var divContentCurrentPage string
	var divContentNextPage string

	// Refresh library list
	minIdx := LibraryPageSize * (c.libraryState.displayedPage - 1)
	if minIdx < 0 {
		minIdx = 0
	}
	maxIdx := LibraryPageSize * (c.libraryState.displayedPage + 2)
	if maxIdx > c.libraryState.cachedSize() {
		maxIdx = c.libraryState.cachedSize()
	}

	step1Idx := LibraryPageSize * c.libraryState.displayedPage
	if step1Idx > maxIdx {
		step1Idx = maxIdx
	}
	step2Idx := LibraryPageSize * (c.libraryState.displayedPage + 1)
	if step2Idx > maxIdx {
		step2Idx = maxIdx
	}

	switch c.libraryState.libraryType {
	case LibraryTypeArtists:
		/*
			for idx, artist := range c.libraryState.cachedArtists[minIdx:maxIdx] {
				if idx < step1Idx {
					c.addArtistItem(&divContentPreviousPage, artist)
				} else if idx < step2Idx {
					c.addArtistItem(&divContentCurrentPage, artist)
				} else {
					c.addArtistItem(&divContentNextPage, artist)
				}
			}
		*/
		divContentPreviousPage = c.renderArtistItemList(c.libraryState.cachedArtists[minIdx:step1Idx])
		divContentCurrentPage = c.renderArtistItemList(c.libraryState.cachedArtists[step1Idx:step2Idx])
		divContentNextPage = c.renderArtistItemList(c.libraryState.cachedArtists[step2Idx:maxIdx])
	case LibraryTypeAlbums:
		/*
			for idx, album := range c.libraryState.cachedAlbums[minIdx:maxIdx] {
				if idx < step1Idx {
					c.addAlbumItem(&divContentPreviousPage, album)
				} else if idx < step2Idx {
					c.addAlbumItem(&divContentCurrentPage, album)
				} else {
					c.addAlbumItem(&divContentNextPage, album)
				}
			}
		*/
		divContentPreviousPage = c.renderAlbumItemList(c.libraryState.cachedAlbums[minIdx:step1Idx])
		divContentCurrentPage = c.renderAlbumItemList(c.libraryState.cachedAlbums[step1Idx:step2Idx])
		divContentNextPage = c.renderAlbumItemList(c.libraryState.cachedAlbums[step2Idx:maxIdx])
	case LibraryTypePlaylists:
		/*
			for idx, playlist := range c.libraryState.cachedPlaylists[minIdx:maxIdx] {
				if idx < step1Idx {
					c.addPlaylistItem(&divContentPreviousPage, playlist)
				} else if idx < step2Idx {
					c.addPlaylistItem(&divContentCurrentPage, playlist)
				} else {
					c.addPlaylistItem(&divContentNextPage, playlist)
				}
			}
		*/
		divContentPreviousPage = c.renderPlaylistItemList(c.libraryState.cachedPlaylists[minIdx:step1Idx])
		divContentCurrentPage = c.renderPlaylistItemList(c.libraryState.cachedPlaylists[step1Idx:step2Idx])
		divContentNextPage = c.renderPlaylistItemList(c.libraryState.cachedPlaylists[step2Idx:maxIdx])
	case LibraryTypeSongs:
		/*
			for idx, song := range c.libraryState.cachedSongs[minIdx:maxIdx] {
				if idx < step1Idx {
					c.addSongItem(&divContentPreviousPage, song)
				} else if idx < step2Idx {
					c.addSongItem(&divContentCurrentPage, song)
				} else {
					c.addSongItem(&divContentNextPage, song)
				}
			}
		*/
		divContentPreviousPage = c.renderSongItemList(c.libraryState.cachedSongs[minIdx:step1Idx])
		divContentCurrentPage = c.renderSongItemList(c.libraryState.cachedSongs[step1Idx:step2Idx])
		divContentNextPage = c.renderSongItemList(c.libraryState.cachedSongs[step2Idx:maxIdx])
	case LibraryTypeUsers:
		/*
			for idx, user := range c.libraryState.cachedUsers {
				if idx < step1Idx {
					c.addUserItem(&divContentPreviousPage, user)
				} else if idx < step2Idx {
					c.addUserItem(&divContentCurrentPage, user)
				} else {
					c.addUserItem(&divContentNextPage, user)
				}
			}
		*/
		divContentPreviousPage = c.renderUserItemList(c.libraryState.cachedUsers[minIdx:step1Idx])
		divContentCurrentPage = c.renderUserItemList(c.libraryState.cachedUsers[step1Idx:step2Idx])
		divContentNextPage = c.renderUserItemList(c.libraryState.cachedUsers[step2Idx:maxIdx])
	}
	var newScrollTop int
	libraryList.Set("innerHTML", divContentPreviousPage)
	if direction == -1 {
		newScrollTop = libraryList.Get("scrollHeight").Int()
	}
	libraryList.Call("insertAdjacentHTML", "beforeEnd", divContentCurrentPage)
	if direction == 1 {
		newScrollTop = libraryList.Get("scrollHeight").Int() - libraryList.Get("clientHeight").Int()
	}
	libraryList.Call("insertAdjacentHTML", "beforeEnd", divContentNextPage)

	if direction != 0 {
		libraryList.Set("scrollTop", newScrollTop)
	}
}

func (c *LibraryComponent) renderArtistItemList(artistList []*restApiV1.Artist) string {
	type ArtistItem struct {
		ArtistId        string
		ArtistName      string
		ArtistSongCount int
		IsEditable      bool
	}

	var artistItemList = make([]ArtistItem, len(artistList))

	for artistIdx, artist := range artistList {
		if artist == nil {
			artistItemList[artistIdx].ArtistId = string(restApiV1.UnknownArtistId)
			artistItemList[artistIdx].ArtistName = "(Unknown artist)"
			artistItemList[artistIdx].ArtistSongCount = len(c.app.localDb.UnknownArtistSongs)
			artistItemList[artistIdx].IsEditable = false
		} else {
			artistItemList[artistIdx].ArtistId = string(artist.Id)
			artistItemList[artistIdx].ArtistName = artist.Name
			artistItemList[artistIdx].ArtistSongCount = len(c.app.localDb.ArtistOrderedSongs[artist.Id])
			artistItemList[artistIdx].IsEditable = c.app.IsConnectedUserAdmin()
		}
	}

	return c.app.RenderTemplate(artistItemList, "home/library/artistItemList")
}

func (c *LibraryComponent) renderAlbumItemList(albumList []*restApiV1.Album) string {
	type AlbumItem struct {
		AlbumId        string
		AlbumName      string
		AlbumSongCount int
		Artists        []struct {
			ArtistId   string
			ArtistName string
		}
		IsEditable bool
	}

	var albumItemList = make([]AlbumItem, len(albumList))

	for albumIdx, album := range albumList {
		if album == nil {
			albumItemList[albumIdx].AlbumId = string(restApiV1.UnknownAlbumId)
			albumItemList[albumIdx].AlbumName = "(Unknown album)"
			albumItemList[albumIdx].AlbumSongCount = len(c.app.localDb.UnknownAlbumSongs)
			albumItemList[albumIdx].IsEditable = false
		} else {
			albumItemList[albumIdx].AlbumId = string(album.Id)
			albumItemList[albumIdx].AlbumName = album.Name
			albumItemList[albumIdx].AlbumSongCount = len(c.app.localDb.AlbumOrderedSongs[album.Id])
			for _, artistId := range album.ArtistIds {
				albumItemList[albumIdx].Artists = append(albumItemList[albumIdx].Artists, struct {
					ArtistId   string
					ArtistName string
				}{
					ArtistId:   string(artistId),
					ArtistName: c.app.localDb.Artists[artistId].Name,
				})
			}
			albumItemList[albumIdx].IsEditable = c.app.IsConnectedUserAdmin()
		}
	}

	return c.app.RenderTemplate(albumItemList, "home/library/albumItemList")
}

func (c *LibraryComponent) renderSongItemList(songList []*restApiV1.Song) string {

	type SongItem struct {
		SongId    string
		Favorite  bool
		SongName  string
		AlbumId   *string
		AlbumName string
		Artists   []struct {
			ArtistId   string
			ArtistName string
		}
		ExplicitFg bool
		IsEditable bool
	}

	var songItemList = make([]SongItem, len(songList))

	for songIdx, song := range songList {
		_, favorite := c.app.localDb.UserFavoriteSongIds[c.app.ConnectedUserId()][song.Id]

		songItemList[songIdx].SongId = string(song.Id)
		songItemList[songIdx].Favorite = favorite
		songItemList[songIdx].SongName = song.Name
		songItemList[songIdx].ExplicitFg = song.ExplicitFg
		songItemList[songIdx].IsEditable = c.app.IsConnectedUserAdmin()

		if song.AlbumId != restApiV1.UnknownAlbumId && c.libraryState.albumId == nil {
			songItemList[songIdx].AlbumName = c.app.localDb.Albums[song.AlbumId].Name
			songItemList[songIdx].AlbumId = (*string)(&song.AlbumId)
		}

		for _, artistId := range song.ArtistIds {
			if c.libraryState.artistId == nil || (c.libraryState.artistId != nil && artistId != *c.libraryState.artistId) {
				songItemList[songIdx].Artists = append(songItemList[songIdx].Artists, struct {
					ArtistId   string
					ArtistName string
				}{
					ArtistId:   string(artistId),
					ArtistName: c.app.localDb.Artists[artistId].Name,
				})
			}
		}
	}

	return c.app.RenderTemplate(songItemList, "home/library/songItemList")
}

func (c *LibraryComponent) renderPlaylistItemList(playlistList []*restApiV1.Playlist) string {
	type PlaylistItem struct {
		PlaylistId        string
		Favorite          bool
		Name              string
		PlaylistSongCount int
		OwnerUsers        []struct {
			UserId   string
			UserName string
		}
		IsEditable  bool
		IsDeletable bool
	}

	var playlistItemList = make([]PlaylistItem, len(playlistList))

	for playlistIdx, playlist := range playlistList {
		_, favorite := c.app.localDb.UserFavoritePlaylistIds[c.app.ConnectedUserId()][playlist.Id]

		playlistItemList[playlistIdx].PlaylistId = string(playlist.Id)
		playlistItemList[playlistIdx].Favorite = favorite
		playlistItemList[playlistIdx].Name = playlist.Name
		playlistItemList[playlistIdx].PlaylistSongCount = len(playlist.SongIds)
		playlistItemList[playlistIdx].IsEditable = c.app.IsConnectedUserAdmin() || c.app.localDb.IsPlaylistOwnedBy(playlist.Id, c.app.ConnectedUserId())
		playlistItemList[playlistIdx].IsDeletable = playlist.Id != restApiV1.IncomingPlaylistId && (c.app.IsConnectedUserAdmin() || c.app.localDb.IsPlaylistOwnedBy(playlist.Id, c.app.ConnectedUserId()))

		for _, userId := range playlist.OwnerUserIds {
			playlistItemList[playlistIdx].OwnerUsers = append(playlistItemList[playlistIdx].OwnerUsers, struct {
				UserId   string
				UserName string
			}{
				UserId:   string(userId),
				UserName: c.app.localDb.Users[userId].Name,
			})
		}

	}

	return c.app.RenderTemplate(playlistItemList, "home/library/playlistItemList")
}

func (c *LibraryComponent) renderUserItemList(userList []*restApiV1.User) string {
	type UserItem struct {
		UserId     string
		Name       string
		IsEditable bool
	}

	var userItemList = make([]UserItem, len(userList))

	for userIdx, user := range userList {
		userItemList[userIdx].UserId = string(user.Id)
		userItemList[userIdx].Name = user.Name
		userItemList[userIdx].IsEditable = c.app.IsConnectedUserAdmin() || c.app.ConnectedUserId() == user.Id
	}

	return c.app.RenderTemplate(userItemList, "home/library/userItemList")
}

func (c *LibraryComponent) ShowArtistsAction() {
	c.libraryState = libraryState{
		libraryType: LibraryTypeArtists,
	}
	jst.Id("librarySearchInput").Set("value", "")
	c.RefreshView()
}

func (c *LibraryComponent) ShowAlbumsAction() {
	c.libraryState = libraryState{
		libraryType: LibraryTypeAlbums,
	}
	jst.Id("librarySearchInput").Set("value", "")
	c.RefreshView()
}

func (c *LibraryComponent) ShowSongsAction() {
	c.libraryState = libraryState{
		libraryType: LibraryTypeSongs,
	}
	jst.Id("librarySearchInput").Set("value", "")
	c.RefreshView()
}

func (c *LibraryComponent) ShowPlaylistsAction() {
	c.libraryState = libraryState{
		libraryType: LibraryTypePlaylists,
	}
	jst.Id("librarySearchInput").Set("value", "")
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
	jst.Id("librarySearchInput").Set("value", "")
	c.RefreshView()
}

func (c *LibraryComponent) OpenArtistAction(artistId restApiV1.ArtistId) {
	c.libraryState = libraryState{
		libraryType: LibraryTypeSongs,
		artistId:    &artistId,
	}
	jst.Id("librarySearchInput").Set("value", "")
	c.RefreshView()
}

func (c *LibraryComponent) OpenPlaylistAction(playlistId restApiV1.PlaylistId) {
	c.libraryState = libraryState{
		libraryType: LibraryTypeSongs,
		playlistId:  &playlistId,
	}
	jst.Id("librarySearchInput").Set("value", "")
	c.RefreshView()
}

func (c *LibraryComponent) FavoritesSwitchAction() {
	c.libraryState.onlyFavoritesFilter = !c.libraryState.onlyFavoritesFilter
	c.RefreshView()
}

func (c *LibraryComponent) SearchAction() {
	librarySearchInput := jst.Id("librarySearchInput")
	nameFilter := librarySearchInput.Get("value").String()

	if nameFilter != "" {
		c.libraryState.nameFilter = &nameFilter
	} else {
		c.libraryState.nameFilter = nil
	}
	c.RefreshView()
}
