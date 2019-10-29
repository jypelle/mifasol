package ui

import (
	"github.com/gdamore/tcell"
	"github.com/jypelle/mifasol/primitive"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/rivo/tview"
	"math/rand"
)

type CurrentComponent struct {
	*tview.Flex
	title *tview.TextView
	list  *primitive.RichList

	songIds       []restApiV1.SongId
	srcPlaylistId *restApiV1.PlaylistId
	modified      bool

	uiApp *UIApp
}

func NewCurrentComponent(uiApp *UIApp) *CurrentComponent {

	c := &CurrentComponent{
		uiApp: uiApp,
	}

	c.title = tview.NewTextView()
	c.title.SetDynamicColors(true)
	c.SetModified(true)

	c.list = primitive.NewRichList()
	c.list.SetInfiniteScroll(false)
	c.list.SetHighlightFullLine(true)
	c.list.SetSelectedBackgroundColor(ColorSelected)
	c.list.SetUnfocusedSelectedBackgroundColor(ColorUnfocusedSelected)
	c.list.SetBorder(false)

	c.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(c.title, 1, 0, false).
		AddItem(c.list, 0, 1, false)

	c.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyRune:
			switch event.Rune() {
			case 's':
				OpenPlaylistContentSaveComponent(c.uiApp, c.songIds, c.srcPlaylistId, c)
			case 'c':
				c.Clear()
				c.SetModified(true)
			case 'r':
				// Shuffle songs list
				rand.Shuffle(len(c.songIds), func(i, j int) {
					c.songIds[i], c.songIds[j] = c.songIds[j], c.songIds[i]
				})
				c.SetModified(true)
				c.RefreshView()

			case 'd':
				if len(c.songIds) > 0 {
					oldIndex := c.list.GetCurrentItem()
					c.list.RemoveItem(oldIndex)
					c.songIds = append(c.songIds[:oldIndex], c.songIds[oldIndex+1:]...)
					c.SetModified(true)
				}
			case '8':
				if len(c.songIds) > 0 {
					srcIndex := c.list.GetCurrentItem()
					if srcIndex > 0 {
						textToMove := c.list.GetItemText(srcIndex - 1)
						songIdToMove := c.songIds[srcIndex-1]

						c.list.RemoveItem(srcIndex - 1)
						c.list.InsertItem(srcIndex, textToMove)

						c.songIds[srcIndex-1] = c.songIds[srcIndex]
						c.songIds[srcIndex] = songIdToMove
						c.SetModified(true)
					}
				}
			case '2':
				if len(c.songIds) > 0 {
					srcIndex := c.list.GetCurrentItem()
					if srcIndex < len(c.songIds)-1 {
						textToMove := c.list.GetItemText(srcIndex + 1)
						songIdToMove := c.songIds[srcIndex+1]

						c.list.RemoveItem(srcIndex + 1)
						c.list.InsertItem(srcIndex, textToMove)

						c.songIds[srcIndex+1] = c.songIds[srcIndex]
						c.songIds[srcIndex] = songIdToMove
						c.SetModified(true)
					}

				}
			}
		case event.Key() == tcell.KeyEnter:
			if len(c.songIds) > 0 {
				songId := c.songIds[c.list.GetCurrentItem()]
				c.uiApp.playerComponent.Play(songId)
				return nil
			}
		}

		return event
	})

	return c
}

func (c *CurrentComponent) Focus(delegate func(tview.Primitive)) {
	delegate(c.list)
}

func (c *CurrentComponent) SetModified(modified bool) {
	c.modified = modified

	title := "[" + ColorTitleStr + "]🎵 Playlist: "

	if c.srcPlaylistId != nil {
		if playlist, ok := c.uiApp.localDb.Playlists[*c.srcPlaylistId]; ok == true {
			title += playlist.Name
		} else {
			title += "(new)"
		}
	} else {
		title += "(new)"
	}

	if c.modified {
		title += " *"
	}
	c.title.SetText(title)
}

func (c *CurrentComponent) Enable() {
	c.title.SetBackgroundColor(ColorTitleBackground)
	c.list.SetBackgroundColor(ColorEnabled)
}

func (c *CurrentComponent) Disable() {
	c.title.SetBackgroundColor(ColorTitleUnfocusedBackground)
	c.list.SetBackgroundColor(ColorDisabled)
}

func (c *CurrentComponent) AddSong(songId restApiV1.SongId) {
	c.songIds = append(c.songIds, songId)
	c.list.AddItem(c.getMainTextSong(songId, -1))
	c.SetModified(true)
}

func (c *CurrentComponent) LoadSong(songId restApiV1.SongId) {
	c.Clear()
	c.AddSong(songId)
}

func (c *CurrentComponent) getMainTextSong(songId restApiV1.SongId, highlightPosition int) string {
	song := c.uiApp.localDb.Songs[songId]

	songName := "[" + ColorSongStr + "]" + tview.Escape(song.Name) + "[white]"

	albumName := ""
	if song.AlbumId != "" {
		albumName = " [::b]/[::-] [" + ColorAlbumStr + "]" + tview.Escape(c.uiApp.localDb.Albums[song.AlbumId].Name) + "[white]"
	}
	artistsName := ""
	if len(song.ArtistIds) > 0 {
		for _, artistId := range song.ArtistIds {
			artistsName += " [::b]/[::-] [" + ColorArtistStr + "]" + tview.Escape(c.uiApp.localDb.Artists[artistId].Name) + "[white]"
		}
	}
	return songName + albumName + artistsName
}

func (c *CurrentComponent) AddSongsFromAlbum(album *restApiV1.Album) {
	if album != nil {
		for _, song := range c.uiApp.localDb.AlbumOrderedSongs[album.Id] {
			c.AddSong(song.Id)
		}
	} else {
		for _, song := range c.uiApp.localDb.UnknownAlbumSongs {
			c.AddSong(song.Id)
		}
	}
}

func (c *CurrentComponent) LoadSongsFromAlbum(album *restApiV1.Album) {
	c.Clear()
	c.SetModified(true)
	c.AddSongsFromAlbum(album)
}

func (c *CurrentComponent) AddSongsFromArtist(artist *restApiV1.Artist) {
	if artist != nil {
		for _, song := range c.uiApp.localDb.ArtistOrderedSongs[artist.Id] {
			c.AddSong(song.Id)
		}
	} else {
		for _, song := range c.uiApp.localDb.UnknownArtistSongs {
			c.AddSong(song.Id)
		}
	}
}

func (c *CurrentComponent) LoadSongsFromArtist(artist *restApiV1.Artist) {
	c.Clear()
	c.SetModified(true)
	c.AddSongsFromArtist(artist)
}

func (c *CurrentComponent) AddSongsFromPlaylist(playlist *restApiV1.Playlist) {
	for _, songId := range playlist.SongIds {
		c.AddSong(songId)
	}
}

func (c *CurrentComponent) LoadSongsFromPlaylist(playlist *restApiV1.Playlist) {
	c.Clear()
	id := playlist.Id
	c.srcPlaylistId = &id
	c.AddSongsFromPlaylist(playlist)
	c.SetModified(false)
}

func (c *CurrentComponent) GetNextSong() *restApiV1.SongId {
	nextPosition := c.list.GetCurrentItem() + 1
	if nextPosition < len(c.songIds) {
		c.list.SetCurrentItem(nextPosition)
		return &c.songIds[nextPosition]
	}

	return nil
}

func (c *CurrentComponent) RefreshView() {
	oldIndex := c.list.GetCurrentItem()
	oldSongIds := c.songIds
	oldSrcPlaylistId := c.srcPlaylistId
	c.Clear()

	// Remove deleted songId
	for _, songId := range oldSongIds {

		if _, ok := c.uiApp.localDb.Songs[songId]; ok {
			c.songIds = append(c.songIds, songId)
			c.list.AddItem(c.getMainTextSong(songId, -1))
		}

	}

	if oldSrcPlaylistId != nil {
		if _, ok := c.uiApp.localDb.Playlists[*oldSrcPlaylistId]; ok {
			c.srcPlaylistId = oldSrcPlaylistId
		} else {
			// If src playlist has been deleted, current playlist is a new playlist
			c.SetModified(true)
		}
	}

	c.list.SetCurrentItem(oldIndex)
}

func (c *CurrentComponent) Clear() {
	c.list.Clear()
	c.songIds = []restApiV1.SongId{}
	c.srcPlaylistId = nil
	c.list.SetCurrentItem(0)
}
