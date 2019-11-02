package ui

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/jypelle/mifasol/primitive"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/rivo/tview"
	"strconv"
)

type LibraryComponent struct {
	*tview.Flex
	title         *tview.TextView
	list          *primitive.RichList
	uiApp         *UIApp
	historyFilter []*libraryFilter
	songs         []*restApiV1.Song
	albums        []*restApiV1.Album
	artists       []*restApiV1.Artist
	playlists     []*restApiV1.Playlist
}

type libraryFilter struct {
	libraryType libraryType
	artistId    *restApiV1.ArtistId
	albumId     *restApiV1.AlbumId
	playlistId  *restApiV1.PlaylistId
	userId      *restApiV1.UserId
	index       int
	position    int
}

type libraryType int64

const (
	libraryTypeMenu libraryType = iota
	libraryTypeArtists
	libraryTypeAlbums
	libraryTypePlaylists
	libraryTypeSongs
	libraryTypeUsers
	libraryTypeSongsFromArtist
	libraryTypeSongsFromAlbum
)

func (l libraryFilter) label() string {
	switch l.libraryType {
	case libraryTypeMenu:
		return "Menu"
	case libraryTypeArtists:
		if l.userId == nil {
			return "All artists"
		} else {
			return "Favorite artists from %s"
		}
	case libraryTypeAlbums:
		if l.userId == nil {
			return "All albums"
		} else {
			return "Favorite albums from %s"
		}
	case libraryTypePlaylists:
		if l.userId == nil {
			return "All playlists"
		} else {
			return "Favorite playlists from %s"
		}
	case libraryTypeSongs:
		if l.userId == nil {
			if l.playlistId == nil {
				return "All songs"
			} else {
				return "Songs from %s playlist"
			}
		} else {
			return "Favorite songs from %s"
		}
	case libraryTypeUsers:
		return "All users"
	case libraryTypeSongsFromArtist:
		if *l.artistId != restApiV1.UnknownArtistId {
			return "Songs from %s artist"
		} else {
			return "Songs from unknown artists"
		}
	case libraryTypeSongsFromAlbum:
		if *l.albumId != restApiV1.UnknownAlbumId {
			return "Songs from %s album"
		} else {
			return "Songs from unknown albums"
		}
	}
	return ""
}

type libraryMenu int64

const (
	libraryMenuMyFavoriteArtists libraryMenu = iota
	libraryMenuMyFavoriteAlbums
	libraryMenuMyFavoritePlaylists
	libraryMenuMyFavoriteSongs
	libraryMenuAllArtists
	libraryMenuAllAlbums
	libraryMenuAllPlaylists
	libraryMenuAllSongs
	libraryMenuAllUsers
)

func (c libraryMenu) label() string {
	switch c {
	case libraryMenuMyFavoriteArtists:
		return "Favorite artists"
	case libraryMenuMyFavoriteAlbums:
		return "Favorite albums"
	case libraryMenuMyFavoritePlaylists:
		return "Favorite playlists"
	case libraryMenuMyFavoriteSongs:
		return "Favorite songs"
	case libraryMenuAllArtists:
		return "All artists"
	case libraryMenuAllAlbums:
		return "All albums"
	case libraryMenuAllPlaylists:
		return "All playlists"
	case libraryMenuAllSongs:
		return "All songs"
	case libraryMenuAllUsers:
		return "All users"
	}
	return ""
}

var libraryMenus = []libraryMenu{
	libraryMenuMyFavoriteArtists,
	libraryMenuMyFavoriteAlbums,
	libraryMenuMyFavoritePlaylists,
	libraryMenuMyFavoriteSongs,
	libraryMenuAllArtists,
	libraryMenuAllAlbums,
	libraryMenuAllPlaylists,
	libraryMenuAllSongs,
	libraryMenuAllUsers,
}

func NewLibraryComponent(uiApp *UIApp) *LibraryComponent {

	c := &LibraryComponent{
		uiApp: uiApp,
	}

	c.title = tview.NewTextView()
	c.title.SetDynamicColors(true)
	c.title.SetText("[" + ColorTitleStr + "]💿 Library")

	c.list = primitive.NewRichList()
	c.list.SetInfiniteScroll(false)
	c.list.SetHighlightFullLine(true)
	c.list.SetPrefixWithLineNumber(true)
	c.list.SetSelectedBackgroundColor(ColorSelected)
	c.list.SetUnfocusedSelectedBackgroundColor(ColorUnfocusedSelected)
	c.list.SetBorder(false)

	c.list.SetHighlightedMainTextFunc(func(index int, mainText string) string {
		switch c.currentFilter().libraryType {
		case libraryTypeMenu:
			return libraryMenus[c.list.GetCurrentItem()].label()
		case libraryTypeArtists:
			return c.getMainTextArtist(c.artists[c.list.GetCurrentItem()], c.currentFilter().position)
		case libraryTypeAlbums:
			return c.getMainTextAlbum(c.albums[c.list.GetCurrentItem()], c.currentFilter().position)
		case libraryTypePlaylists:
			return c.getMainTextPlaylist(c.playlists[c.list.GetCurrentItem()], nil, c.currentFilter().position)
		case libraryTypeUsers:
			return c.getMainTextUser(c.uiApp.LocalDb().OrderedUsers[c.list.GetCurrentItem()])
		case libraryTypeSongsFromArtist:
			return c.getMainTextSong(c.songs[c.list.GetCurrentItem()], nil, c.currentFilter().artistId, c.currentFilter().position)
		case libraryTypeSongsFromAlbum:
			return c.getMainTextSong(c.songs[c.list.GetCurrentItem()], c.currentFilter().albumId, nil, c.currentFilter().position)
		case libraryTypeSongs:
			return c.getMainTextSong(c.songs[c.list.GetCurrentItem()], nil, nil, c.currentFilter().position)
		}
		return ""
	})
	c.list.SetChangedFunc(func(index int, mainText string) {
		c.currentFilter().index = index
		c.currentFilter().position = 0
	})

	c.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(c.title, 1, 0, false).
		AddItem(c.list, 0, 1, false)

	c.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentFilter := c.currentFilter()

		switch {

		case event.Key() == tcell.KeyRune:
			switch event.Rune() {
			case 'a':
				if c.list.GetItemCount() > 0 && currentFilter.libraryType != libraryTypeMenu && currentFilter.libraryType != libraryTypeUsers {
					switch currentFilter.libraryType {
					case libraryTypeArtists:
						artist := c.artists[c.list.GetCurrentItem()]
						c.uiApp.CurrentComponent().AddSongsFromArtist(artist)
					case libraryTypeAlbums:
						album := c.albums[c.list.GetCurrentItem()]
						c.uiApp.CurrentComponent().AddSongsFromAlbum(album)
					case libraryTypePlaylists:
						playlist := c.playlists[c.list.GetCurrentItem()]
						c.uiApp.CurrentComponent().AddSongsFromPlaylist(playlist)
					case libraryTypeSongs,
						libraryTypeSongsFromAlbum,
						libraryTypeSongsFromArtist:
						song := c.songs[c.list.GetCurrentItem()]
						c.uiApp.CurrentComponent().AddSong(song.Id)
					}
					c.list.SetCurrentItem(c.list.GetCurrentItem() + 1)
				}
				return nil

			case 'l':
				if c.list.GetItemCount() > 0 && currentFilter.libraryType != libraryTypeMenu && currentFilter.libraryType != libraryTypeUsers {
					switch currentFilter.libraryType {
					case libraryTypeArtists:
						artist := c.artists[c.list.GetCurrentItem()]
						c.uiApp.CurrentComponent().LoadSongsFromArtist(artist)
					case libraryTypeAlbums:
						album := c.albums[c.list.GetCurrentItem()]
						c.uiApp.CurrentComponent().LoadSongsFromAlbum(album)
					case libraryTypePlaylists:
						playlist := c.playlists[c.list.GetCurrentItem()]
						c.uiApp.CurrentComponent().LoadSongsFromPlaylist(playlist)
					case libraryTypeSongs,
						libraryTypeSongsFromArtist,
						libraryTypeSongsFromAlbum:
						song := c.songs[c.list.GetCurrentItem()]
						c.uiApp.CurrentComponent().LoadSong(song.Id)
					}
				}
				return nil

			case 'c':
				if currentFilter.libraryType != libraryTypeMenu {
					switch currentFilter.libraryType {
					case libraryTypeArtists:
						OpenArtistCreateComponent(c.uiApp, c)
					case libraryTypeAlbums:
						OpenAlbumCreateComponent(c.uiApp, c)
					case libraryTypeUsers:
						OpenUserCreateComponent(c.uiApp, c)
					}
				}
				return nil

			case 'd':
				if c.list.GetItemCount() > 0 && currentFilter.libraryType != libraryTypeMenu {
					switch currentFilter.libraryType {
					case libraryTypeArtists:
						artist := c.artists[c.list.GetCurrentItem()]
						if artist != nil {
							c.uiApp.ConfirmArtistDelete(artist)
						}
					case libraryTypeAlbums:
						album := c.albums[c.list.GetCurrentItem()]
						if album != nil {
							c.uiApp.ConfirmAlbumDelete(album)
						}
					case libraryTypePlaylists:
						playlist := c.playlists[c.list.GetCurrentItem()]
						if playlist != nil {
							c.uiApp.ConfirmPlaylistDelete(playlist)
						}
					case libraryTypeUsers:
						user := c.uiApp.LocalDb().OrderedUsers[c.list.GetCurrentItem()]
						if user != nil {
							c.uiApp.ConfirmUserDelete(user)
						}
					case libraryTypeSongs,
						libraryTypeSongsFromAlbum,
						libraryTypeSongsFromArtist:
						song := c.songs[c.list.GetCurrentItem()]
						if song != nil {
							c.uiApp.ConfirmSongDelete(song)
						}
					}
				}
				return nil

			case 'e':
				if c.list.GetItemCount() > 0 && currentFilter.libraryType != libraryTypeMenu {
					switch currentFilter.libraryType {
					case libraryTypeArtists:
						artist := c.artists[c.list.GetCurrentItem()]
						if artist != nil {
							OpenArtistEditComponent(c.uiApp, artist.Id, &artist.ArtistMeta, c)
						}
					case libraryTypeAlbums:
						album := c.albums[c.list.GetCurrentItem()]
						if album != nil {
							OpenAlbumEditComponent(c.uiApp, album.Id, &album.AlbumMeta, c)
						}
					case libraryTypePlaylists:
						playlist := c.playlists[c.list.GetCurrentItem()]
						if playlist != nil {
							OpenPlaylistEditComponent(c.uiApp, playlist, c)
						}
					case libraryTypeUsers:
						user := c.uiApp.LocalDb().OrderedUsers[c.list.GetCurrentItem()]
						if user != nil {
							OpenUserEditComponent(c.uiApp, user.Id, &user.UserMeta, c)
						}
					case libraryTypeSongs,
						libraryTypeSongsFromAlbum,
						libraryTypeSongsFromArtist:
						song := c.songs[c.list.GetCurrentItem()]
						if song != nil {
							OpenSongEditComponent(c.uiApp, song, c)
						}
					}
				}
				return nil

			case 'f':
				if c.list.GetItemCount() > 0 && currentFilter.libraryType != libraryTypeMenu {
					switch currentFilter.libraryType {
					case libraryTypePlaylists:
						playlist := c.playlists[c.list.GetCurrentItem()]
						if playlist != nil {
							myFavoritePlaylistIds := c.uiApp.LocalDb().UserFavoritePlaylistIds[c.uiApp.ConnectedUserId()]
							favoritePlaylistId := restApiV1.FavoritePlaylistId{
								UserId:     c.uiApp.ConnectedUserId(),
								PlaylistId: playlist.Id,
							}
							if _, ok := myFavoritePlaylistIds[playlist.Id]; ok {
								_, cliErr := c.uiApp.restClient.DeleteFavoritePlaylist(favoritePlaylistId)
								if cliErr != nil {
									c.uiApp.ClientErrorMessage("Unable to add playlist to favorites", cliErr)
								}
								c.uiApp.Reload()
							} else {
								_, cliErr := c.uiApp.restClient.CreateFavoritePlaylist(&restApiV1.FavoritePlaylistMeta{Id: favoritePlaylistId})
								if cliErr != nil {
									c.uiApp.ClientErrorMessage("Unable to remove playlist from favorites", cliErr)
								}
								c.uiApp.Reload()
							}
							if !(currentFilter.userId != nil && *currentFilter.userId == c.uiApp.ConnectedUserId()) {
								c.list.SetCurrentItem(c.list.GetCurrentItem() + 1)
							}
						}
					case libraryTypeSongs,
						libraryTypeSongsFromAlbum,
						libraryTypeSongsFromArtist:
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
								c.RefreshView()
							} else {
								_, cliErr := c.uiApp.restClient.CreateFavoriteSong(&restApiV1.FavoriteSongMeta{Id: favoriteSongId})
								if cliErr != nil {
									c.uiApp.ClientErrorMessage("Unable to remove song from favorites", cliErr)
								}
								c.uiApp.LocalDb().AddSongToMyFavorite(song.Id)
								c.RefreshView()
							}
							if !(currentFilter.userId != nil && *currentFilter.userId == c.uiApp.ConnectedUserId()) {
								c.list.SetCurrentItem(c.list.GetCurrentItem() + 1)
							}
						}
					}
				}
				return nil
			case '/':
				if c.list.GetItemCount() > 0 && currentFilter.libraryType != libraryTypeMenu {
					switch currentFilter.libraryType {
					case libraryTypeSongs,
						libraryTypeSongsFromAlbum,
						libraryTypeSongsFromArtist:
					}
				}
				return nil

			}
		case event.Key() == tcell.KeyDEL, event.Key() == tcell.KeyBackspace:
			c.GoToPreviousFilter()
			return nil
		case event.Key() == tcell.KeyLeft:
			if currentFilter.position > 0 {
				currentFilter.position--
			}
		case event.Key() == tcell.KeyRight:
			if currentFilter.position < 30 {
				currentFilter.position++
			}
		case event.Key() == tcell.KeyEnter:
			if c.list.GetItemCount() > 0 {
				switch currentFilter.libraryType {
				case libraryTypeMenu:
					libraryMenu := libraryMenus[c.list.GetCurrentItem()]
					switch libraryMenu {
					case libraryMenuMyFavoriteArtists:
						c.GoToFavoriteArtistsFromUserFilter(c.uiApp.ConnectedUserId())
					case libraryMenuMyFavoriteAlbums:
						c.GoToFavoriteAlbumsFromUserFilter(c.uiApp.ConnectedUserId())
					case libraryMenuMyFavoritePlaylists:
						c.GoToFavoritePlaylistsFromUserFilter(c.uiApp.ConnectedUserId())
					case libraryMenuMyFavoriteSongs:
						c.GoToFavoriteSongsFromUserFilter(c.uiApp.ConnectedUserId())
					case libraryMenuAllArtists:
						c.GoToAllArtistsFilter()
					case libraryMenuAllAlbums:
						c.GoToAllAlbumsFilter()
					case libraryMenuAllPlaylists:
						c.GoToAllPlaylistsFilter()
					case libraryMenuAllSongs:
						c.GoToAllSongsFilter()
					case libraryMenuAllUsers:
						c.GoToAllUsersFilter()
					}
				case libraryTypeArtists:
					artist := c.artists[c.list.GetCurrentItem()]
					if artist == nil {
						c.GoToSongsFromUnknownArtistFilter()
					} else {
						songId, artistId, albumId := c.getPositionnedIdArtist(c.artists[c.list.GetCurrentItem()], c.currentFilter().position)
						c.open(songId, artistId, albumId)
					}
				case libraryTypeAlbums:
					album := c.albums[c.list.GetCurrentItem()]
					if album == nil {
						c.GoToSongsFromUnknownAlbumFilter()
					} else {
						songId, artistId, albumId := c.getPositionnedIdAlbum(c.albums[c.list.GetCurrentItem()], c.currentFilter().position)
						c.open(songId, artistId, albumId)
					}
				case libraryTypePlaylists:
					playlist := c.playlists[c.list.GetCurrentItem()]
					c.GoToSongsFromPlaylistFilter(playlist.Id)
				case libraryTypeSongs,
					libraryTypeSongsFromAlbum,
					libraryTypeSongsFromArtist:

					songId, artistId, albumId := c.getPositionnedIdSong(c.songs[c.list.GetCurrentItem()], c.currentFilter().albumId, c.currentFilter().artistId, c.currentFilter().position)
					c.open(songId, artistId, albumId)
				case libraryTypeUsers:
				}
			}
			return nil

		}

		return event
	})

	// Start with menu
	c.ResetToMenuFilter()

	return c
}

func (c *LibraryComponent) Enable() {
	c.title.SetBackgroundColor(ColorTitleBackground)
	c.list.SetBackgroundColor(ColorEnabled)
}

func (c *LibraryComponent) Disable() {
	c.title.SetBackgroundColor(ColorTitleUnfocusedBackground)
	c.list.SetBackgroundColor(ColorDisabled)
}

func (c *LibraryComponent) getTitlePrefix() string {
	return "[" + ColorTitleStr + "]💿 Library"
}

func (c *LibraryComponent) currentFilter() *libraryFilter {
	return c.historyFilter[len(c.historyFilter)-1]
}

func (c *LibraryComponent) historizeLibraryFilter(newLibraryFilter *libraryFilter) {
	if len(c.historyFilter) > 0 {
		c.currentFilter().index = c.list.GetCurrentItem()
	}
	c.historyFilter = append(c.historyFilter, newLibraryFilter)
	c.RefreshView()
}

func (c *LibraryComponent) GoToPreviousFilter() {
	if len(c.historyFilter) > 1 {
		c.historyFilter = c.historyFilter[:len(c.historyFilter)-1]
	}
	c.RefreshView()
}

func (c *LibraryComponent) ResetToMenuFilter() {
	c.historyFilter = nil
	c.GoToMenuFilter()
}

func (c *LibraryComponent) GoToMenuFilter() {
	c.historizeLibraryFilter(&libraryFilter{libraryType: libraryTypeMenu})
}

func (c *LibraryComponent) GoToAllArtistsFilter() {
	c.historizeLibraryFilter(&libraryFilter{libraryType: libraryTypeArtists})
}

func (c *LibraryComponent) GoToAllAlbumsFilter() {
	c.historizeLibraryFilter(&libraryFilter{libraryType: libraryTypeAlbums})
}

func (c *LibraryComponent) GoToAllPlaylistsFilter() {
	c.historizeLibraryFilter(&libraryFilter{libraryType: libraryTypePlaylists})
}

func (c *LibraryComponent) GoToAllSongsFilter() {
	c.historizeLibraryFilter(&libraryFilter{libraryType: libraryTypeSongs})
}

func (c *LibraryComponent) GoToAllUsersFilter() {
	c.historizeLibraryFilter(&libraryFilter{libraryType: libraryTypeUsers})
}

func (c *LibraryComponent) GoToSongsFromAlbumFilter(albumId restApiV1.AlbumId) {
	c.historizeLibraryFilter(&libraryFilter{libraryType: libraryTypeSongsFromAlbum, albumId: &albumId})
}

func (c *LibraryComponent) GoToSongsFromUnknownAlbumFilter() {
	c.historizeLibraryFilter(&libraryFilter{libraryType: libraryTypeSongsFromAlbum, albumId: &restApiV1.UnknownAlbumId})
}

func (c *LibraryComponent) GoToSongsFromArtistFilter(artistId restApiV1.ArtistId) {
	c.historizeLibraryFilter(&libraryFilter{libraryType: libraryTypeSongsFromArtist, artistId: &artistId})
}

func (c *LibraryComponent) GoToSongsFromUnknownArtistFilter() {
	c.historizeLibraryFilter(&libraryFilter{libraryType: libraryTypeSongsFromArtist, artistId: &restApiV1.UnknownArtistId})
}

func (c *LibraryComponent) GoToSongsFromPlaylistFilter(playlistId restApiV1.PlaylistId) {
	c.historizeLibraryFilter(&libraryFilter{libraryType: libraryTypeSongs, playlistId: &playlistId})
}

func (c *LibraryComponent) GoToFavoriteArtistsFromUserFilter(userId restApiV1.UserId) {
	c.historizeLibraryFilter(&libraryFilter{libraryType: libraryTypeArtists, userId: &userId})
}

func (c *LibraryComponent) GoToFavoriteAlbumsFromUserFilter(userId restApiV1.UserId) {
	c.historizeLibraryFilter(&libraryFilter{libraryType: libraryTypeAlbums, userId: &userId})
}

func (c *LibraryComponent) GoToFavoritePlaylistsFromUserFilter(userId restApiV1.UserId) {
	c.historizeLibraryFilter(&libraryFilter{libraryType: libraryTypePlaylists, userId: &userId})
}

func (c *LibraryComponent) GoToFavoriteSongsFromUserFilter(userId restApiV1.UserId) {
	c.historizeLibraryFilter(&libraryFilter{libraryType: libraryTypeSongs, userId: &userId})
}

func (c *LibraryComponent) RefreshView() {
	// Redirection to menu when filter references obsolete artist/album/playlist id
	currentFilter := c.currentFilter()

	if currentFilter.albumId != nil && *currentFilter.albumId != restApiV1.UnknownAlbumId {
		if _, ok := c.uiApp.LocalDb().Albums[*currentFilter.albumId]; !ok {
			c.ResetToMenuFilter()
		}
	}
	if currentFilter.artistId != nil && *currentFilter.artistId != restApiV1.UnknownArtistId {
		if _, ok := c.uiApp.LocalDb().Artists[*currentFilter.artistId]; !ok {
			c.ResetToMenuFilter()
		}
	}
	if currentFilter.playlistId != nil {
		if _, ok := c.uiApp.LocalDb().Playlists[*currentFilter.playlistId]; !ok {
			c.ResetToMenuFilter()
		}
	}
	if currentFilter.userId != nil {
		if _, ok := c.uiApp.LocalDb().Users[*currentFilter.userId]; !ok {
			c.ResetToMenuFilter()
		}
	}

	currentFilter = c.currentFilter()
	c.list.Clear()
	oldIndex := currentFilter.index
	c.songs = nil
	title := c.getTitlePrefix() + ": " + currentFilter.label()
	switch currentFilter.libraryType {
	case libraryTypeMenu:
		for _, libraryMenu := range libraryMenus {
			c.list.AddItem(libraryMenu.label())
		}
	case libraryTypeArtists:
		if currentFilter.userId == nil {
			c.artists = c.uiApp.LocalDb().OrderedArtists
		} else {
			user := c.uiApp.LocalDb().Users[*currentFilter.userId]
			title = fmt.Sprintf(title, user.Name)
			c.artists = c.uiApp.LocalDb().UserOrderedFavoriteArtists[*currentFilter.userId]
		}
		for _, artist := range c.artists {
			c.list.AddItem(c.getMainTextArtist(artist, -1))
		}
	case libraryTypeAlbums:
		if currentFilter.userId == nil {
			c.albums = c.uiApp.LocalDb().OrderedAlbums
		} else {
			user := c.uiApp.LocalDb().Users[*currentFilter.userId]
			title = fmt.Sprintf(title, user.Name)
			c.albums = c.uiApp.LocalDb().UserOrderedFavoriteAlbums[*currentFilter.userId]
		}
		for _, album := range c.albums {
			c.list.AddItem(c.getMainTextAlbum(album, -1))
		}
	case libraryTypePlaylists:
		if currentFilter.userId == nil {
			c.playlists = c.uiApp.LocalDb().OrderedPlaylists
		} else {
			user := c.uiApp.LocalDb().Users[*currentFilter.userId]
			title = fmt.Sprintf(title, user.Name)
			c.playlists = c.uiApp.LocalDb().UserOrderedFavoritePlaylists[*currentFilter.userId]
		}
		c.loadPlaylists(c.playlists, nil)
	case libraryTypeSongs:
		if currentFilter.userId == nil {
			if currentFilter.playlistId == nil {
				c.songs = c.uiApp.LocalDb().OrderedSongs
			} else {
				playlist := c.uiApp.LocalDb().Playlists[*currentFilter.playlistId]
				title = fmt.Sprintf(title, playlist.Name)
				songIds := c.uiApp.LocalDb().Playlists[playlist.Id].SongIds
				for _, songId := range songIds {
					c.songs = append(c.songs, c.uiApp.LocalDb().Songs[songId])
				}
			}
		} else {
			user := c.uiApp.LocalDb().Users[*currentFilter.userId]
			title = fmt.Sprintf(title, user.Name)
			c.songs = c.uiApp.LocalDb().UserOrderedFavoriteSongs[*currentFilter.userId]
		}
		c.loadSongs(c.songs, nil, nil)
	case libraryTypeUsers:
		for _, user := range c.uiApp.LocalDb().OrderedUsers {
			c.list.AddItem(c.getMainTextUser(user))
		}
	case libraryTypeSongsFromAlbum:
		if *currentFilter.albumId != restApiV1.UnknownAlbumId {
			album := c.uiApp.LocalDb().Albums[*currentFilter.albumId]
			title = fmt.Sprintf(title, album.Name)
			c.songs = c.uiApp.LocalDb().AlbumOrderedSongs[album.Id]
		} else {
			c.songs = c.uiApp.LocalDb().UnknownAlbumSongs
		}
		c.loadSongs(c.songs, currentFilter.albumId, nil)
	case libraryTypeSongsFromArtist:
		if *currentFilter.artistId != restApiV1.UnknownArtistId {
			artist := c.uiApp.LocalDb().Artists[*currentFilter.artistId]
			title = fmt.Sprintf(title, artist.Name)
			c.songs = c.uiApp.LocalDb().ArtistOrderedSongs[artist.Id]
		} else {
			c.songs = c.uiApp.LocalDb().UnknownArtistSongs
		}
		c.loadSongs(c.songs, nil, currentFilter.artistId)
	}
	c.title.SetText(title)
	c.list.SetCurrentItem(oldIndex)
}

func (c *LibraryComponent) loadSongs(songs []*restApiV1.Song, fromAlbumId *restApiV1.AlbumId, fromArtistId *restApiV1.ArtistId) {
	for _, song := range songs {
		c.list.AddItem(c.getMainTextSong(song, fromAlbumId, fromArtistId, -1))
	}
}

func (c *LibraryComponent) getPositionnedIdSong(song *restApiV1.Song, fromAlbumId *restApiV1.AlbumId, fromArtistId *restApiV1.ArtistId, highlightPosition int) (songId *restApiV1.SongId, artistId *restApiV1.ArtistId, albumId *restApiV1.AlbumId) {
	currentPosition := 0

	// Song name
	if currentPosition == highlightPosition {
		return &song.Id, nil, nil
	}
	currentPosition++

	// Album name
	if song.AlbumId != restApiV1.UnknownAlbumId && fromAlbumId == nil {
		if currentPosition == highlightPosition {
			return nil, nil, &song.AlbumId
		}
		currentPosition++
	}

	if len(song.ArtistIds) > 0 {
		for _, artistId := range song.ArtistIds {
			if fromArtistId == nil || (fromArtistId != nil && artistId != *fromArtistId) {
				if currentPosition == highlightPosition {
					return nil, &artistId, nil
				}
				currentPosition++
			}
		}
	}

	return nil, nil, nil
}

func (c *LibraryComponent) getMainTextSong(song *restApiV1.Song, fromAlbumId *restApiV1.AlbumId, fromArtistId *restApiV1.ArtistId, highlightPosition int) string {

	currentPosition := 0
	text := ""

	myFavoriteSongIds := c.uiApp.LocalDb().UserFavoriteSongIds[c.uiApp.ConnectedUserId()]
	if _, ok := myFavoriteSongIds[song.Id]; ok {
		text += "■ "
	} else {
		text += "  "
	}

	// Song name
	if currentPosition >= highlightPosition {
		underline := ""
		if currentPosition == highlightPosition {
			underline = "u"
		}
		text += "[" + ColorSongStr + "::" + underline + "]" + tview.Escape(song.Name) + "[white::-]"
	}
	currentPosition++

	// Album name
	if song.AlbumId != restApiV1.UnknownAlbumId && fromAlbumId == nil {
		if currentPosition >= highlightPosition {
			if currentPosition > highlightPosition {
				text += " [::b]/[::-] "
			}
			underline := ""
			if currentPosition == highlightPosition {
				underline = "u"
			}
			text += "[" + ColorAlbumStr + "::" + underline + "]" + tview.Escape(c.uiApp.LocalDb().Albums[song.AlbumId].Name) + "[white::-]"
		}
		currentPosition++
	}

	if len(song.ArtistIds) > 0 {
		for _, artistId := range song.ArtistIds {
			if fromArtistId == nil || (fromArtistId != nil && artistId != *fromArtistId) {
				if currentPosition >= highlightPosition {
					if currentPosition > highlightPosition {
						text += " [::b]/[::-] "
					}
					underline := ""
					if currentPosition == highlightPosition {
						underline = "u"
					}
					text += "[" + ColorArtistStr + "::" + underline + "]" + tview.Escape(c.uiApp.LocalDb().Artists[artistId].Name) + "[white::-]"
				}
				currentPosition++
			}
		}
	}

	return text
}

func (c *LibraryComponent) getPositionnedIdAlbum(album *restApiV1.Album, highlightPosition int) (songId *restApiV1.SongId, artistId *restApiV1.ArtistId, albumId *restApiV1.AlbumId) {
	currentPosition := 0

	if album == nil {
		if currentPosition >= highlightPosition {
			return nil, nil, nil
		}
		currentPosition++
	} else {
		if currentPosition >= highlightPosition {
			return nil, nil, &album.Id
		}
		currentPosition++

		if len(album.ArtistIds) > 0 {
			for _, artistId := range album.ArtistIds {
				if currentPosition >= highlightPosition {
					return nil, &artistId, nil
				}
				currentPosition++
			}
		}
	}

	return nil, nil, nil
}

func (c *LibraryComponent) getMainTextAlbum(album *restApiV1.Album, highlightPosition int) string {
	currentPosition := 0
	text := ""

	if album == nil {
		if currentPosition >= highlightPosition {
			text += "[white]" + tview.Escape("(Unknown album)") + "[white] (" + strconv.Itoa(len(c.uiApp.LocalDb().UnknownAlbumSongs)) + ")"
		}
		currentPosition++
	} else {
		if currentPosition >= highlightPosition {
			text += "[" + ColorAlbumStr + "]" + tview.Escape(album.Name) + "[white] (" + strconv.Itoa(len(c.uiApp.LocalDb().AlbumOrderedSongs[album.Id])) + ")"
		}
		currentPosition++

		if len(album.ArtistIds) > 0 {
			for _, artistId := range album.ArtistIds {
				if currentPosition >= highlightPosition {
					if currentPosition > highlightPosition {
						text += " [::b]/[::-] "
					}
					text += "[" + ColorArtistStr + "]" + tview.Escape(c.uiApp.LocalDb().Artists[artistId].Name) + "[white]"
				}
				currentPosition++
			}
		}
	}

	return text
}

func (c *LibraryComponent) getPositionnedIdArtist(artist *restApiV1.Artist, highlightPosition int) (songId *restApiV1.SongId, artistId *restApiV1.ArtistId, albumId *restApiV1.AlbumId) {
	currentPosition := 0

	if artist == nil {
		if currentPosition >= highlightPosition {
			return nil, nil, nil
		}
		currentPosition++
	} else {
		if currentPosition >= highlightPosition {
			return nil, &artist.Id, nil
		}
		currentPosition++
	}

	return nil, nil, nil
}

func (c *LibraryComponent) getMainTextArtist(artist *restApiV1.Artist, highlightPosition int) string {
	currentPosition := 0
	text := ""

	if artist == nil {
		if currentPosition >= highlightPosition {
			text += "[white]" + tview.Escape("(Unknown artist)") + "[white] (" + strconv.Itoa(len(c.uiApp.LocalDb().UnknownArtistSongs)) + ")"
		}
		currentPosition++
	} else {
		if currentPosition >= highlightPosition {
			text += "[" + ColorArtistStr + "]" + tview.Escape(artist.Name) + "[white] (" + strconv.Itoa(len(c.uiApp.LocalDb().ArtistOrderedSongs[artist.Id])) + ")"
		}
		currentPosition++
	}

	return text
}

func (c *LibraryComponent) loadPlaylists(playlists []*restApiV1.Playlist, fromOwnerUserId *restApiV1.UserId) {
	for _, playlist := range playlists {
		c.list.AddItem(c.getMainTextPlaylist(playlist, fromOwnerUserId, -1))
	}
}

func (c *LibraryComponent) getMainTextPlaylist(playlist *restApiV1.Playlist, fromOwnerUserId *restApiV1.UserId, highlightPosition int) string {
	currentPosition := 0
	text := ""

	myFavoritePlaylistIds := c.uiApp.LocalDb().UserFavoritePlaylistIds[c.uiApp.ConnectedUserId()]
	if _, ok := myFavoritePlaylistIds[playlist.Id]; ok {
		text += "■ "
		//text += "💙"
	} else {
		text += "  "
	}

	if currentPosition >= highlightPosition {
		text += "[" + ColorPlaylistStr + "]" + tview.Escape(playlist.Name) + "[white] (" + strconv.Itoa(len(c.uiApp.LocalDb().Playlists[playlist.Id].SongIds)) + ")"
	}
	currentPosition++

	if len(playlist.OwnerUserIds) > 0 {
		for _, userId := range playlist.OwnerUserIds {
			if fromOwnerUserId == nil || (fromOwnerUserId != nil && userId != *fromOwnerUserId) {
				if currentPosition >= highlightPosition {
					if currentPosition > highlightPosition {
						text += " [::b]/[::-] "
					}
					text += "[" + ColorUserStr + "]" + tview.Escape(c.uiApp.LocalDb().Users[userId].Name) + "[white]"
				}
				currentPosition++
			}
		}
	}

	return text
}

func (c *LibraryComponent) getMainTextUser(user *restApiV1.User) string {

	userName := "[" + ColorUserStr + "]" + tview.Escape(user.Name)

	return userName
}

func (c *LibraryComponent) Focus(delegate func(tview.Primitive)) {
	delegate(c.list)
}

func (c *LibraryComponent) open(songId *restApiV1.SongId, artistId *restApiV1.ArtistId, albumId *restApiV1.AlbumId) {
	if songId != nil {
		c.uiApp.Play(*songId)
	}
	if artistId != nil {
		c.GoToSongsFromArtistFilter(*artistId)
	}
	if albumId != nil {
		c.GoToSongsFromAlbumFilter(*albumId)
	}
}
