package cliwa

import (
	"fmt"
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restApiV1"
	"html"
	"math/rand"
	"strconv"
	"strings"
	"syscall/js"
)

type CurrentComponent struct {
	app *App

	songIds        []restApiV1.SongId
	currentSongIdx int
	srcPlaylistId  *restApiV1.PlaylistId
	modified       bool
}

func NewCurrentComponent(app *App) *CurrentComponent {
	c := &CurrentComponent{
		app:      app,
		modified: true,
	}

	return c
}

func (c *CurrentComponent) Show() {
	currentCleanButton := jst.Document.Call("getElementById", "currentCleanButton")
	currentCleanButton.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		c.songIds = nil
		c.srcPlaylistId = nil
		c.modified = true
		c.RefreshView()
	}))
	currentShuffleButton := jst.Document.Call("getElementById", "currentShuffleButton")
	currentShuffleButton.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		rand.Shuffle(len(c.songIds), func(i, j int) { c.songIds[i], c.songIds[j] = c.songIds[j], c.songIds[i] })
		c.modified = true
		c.RefreshView()
	}))
	currentSaveButton := jst.Document.Call("getElementById", "currentSaveButton")
	currentSaveButton.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		if c.modified {
			if c.srcPlaylistId == nil {
				// Save as
				// OpenPlaylistContentSaveComponent(c.uiApp, c.songIds, c.srcPlaylistId, c)
			} else {
				// Save
				// Only admin or playlist owner can edit playlist content
				if !c.app.IsConnectedUserAdmin() && !c.app.localDb.IsPlaylistOwnedBy(*c.srcPlaylistId, c.app.ConnectedUserId()) {
					c.app.HomeComponent.MessageComponent.WarningMessage("Only administrator or playlist owner can edit playlist content")
				} else {
					selectedPlaylist := c.app.localDb.Playlists[*c.srcPlaylistId]
					playlistMeta := selectedPlaylist.PlaylistMeta
					playlistMeta.SongIds = c.songIds

					_, cliErr := c.app.restClient.UpdatePlaylist(selectedPlaylist.Id, &playlistMeta)
					if cliErr != nil {
						c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to update the playlist", cliErr)
					} else {
						var id = selectedPlaylist.Id
						c.srcPlaylistId = &id
						c.modified = false
						c.app.HomeComponent.Reload()
					}

				}
			}
		}
	}))
	currentSaveAsButton := jst.Document.Call("getElementById", "currentSaveAsButton")
	currentSaveAsButton.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		// Save as
		// OpenPlaylistContentSaveComponent(c.uiApp, c.songIds, c.srcPlaylistId, c)
	}))

	listDiv := jst.Document.Call("getElementById", "currentList")
	listDiv.Call("addEventListener", "click", c.app.AddRichEventFunc(func(this js.Value, i []js.Value) {
		link := i[0].Get("target").Call("closest", ".artistLink, .albumLink, .currentPlaySongNowLink, .currentRemoveSongFromPlaylistLink")
		if !link.Truthy() {
			return
		}
		dataset := link.Get("dataset")

		switch link.Get("className").String() {
		case "artistLink":
			artistId := dataset.Get("artistid").String()
			c.app.HomeComponent.LibraryComponent.OpenArtistAction(restApiV1.ArtistId(artistId))
		case "albumLink":
			albumId := dataset.Get("albumid").String()
			c.app.HomeComponent.LibraryComponent.OpenAlbumAction(restApiV1.AlbumId(albumId))
		case "currentPlaySongNowLink":
			songId := dataset.Get("songid").String()
			c.currentSongIdx, _ = strconv.Atoi(dataset.Get("songidx").String())
			c.app.HomeComponent.PlayerComponent.PlaySongAction(restApiV1.SongId(songId))
		case "currentRemoveSongFromPlaylistLink":
			songIdx, _ := strconv.Atoi(dataset.Get("songidx").String())
			c.RemoveSongFromPlaylistAction(songIdx)
		}
	}))

}

func (c *CurrentComponent) PlayNextSongAction() {
	if c.currentSongIdx < len(c.songIds)-1 {
		c.currentSongIdx++
		c.app.HomeComponent.PlayerComponent.PlaySongAction(c.songIds[c.currentSongIdx])
	}
}

func (c *CurrentComponent) RefreshView() {
	listDiv := jst.Document.Call("getElementById", "currentList")

	listDiv.Set("innerHTML", "")
	for songIdx, songId := range c.songIds {
		listDiv.Call("insertAdjacentHTML", "beforeEnd", c.addSongItem(songIdx, c.app.localDb.Songs[songId]))
	}

	// Refresh current playlist title
	titleSpan := jst.Document.Call("getElementById", "currentTitle")
	var title string
	if c.srcPlaylistId == nil {
		title = "New playlist"
	} else {
		title = fmt.Sprintf(`<span class="playlistLink">%s</span>`, html.EscapeString(c.app.localDb.Playlists[*c.srcPlaylistId].Name))
	}
	if c.modified {
		title += " *"
	}
	titleSpan.Set("innerHTML", title)

}

func (c *CurrentComponent) addSongItem(songIdx int, song *restApiV1.Song) string {
	var divContent strings.Builder

	songItem := struct {
		SongId    string
		SongIdx   int
		SongName  string
		AlbumId   *string
		AlbumName string
		Artists   []struct {
			ArtistId   string
			ArtistName string
		}
	}{
		SongId:   string(song.Id),
		SongIdx:  songIdx,
		SongName: song.Name,
	}

	if song.AlbumId != restApiV1.UnknownAlbumId {
		songItem.AlbumName = c.app.localDb.Albums[song.AlbumId].Name
		songItem.AlbumId = (*string)(&song.AlbumId)
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
		&songItem, "currentSongItem.html"),
	)

	return divContent.String()
}

func (c *CurrentComponent) AddSongAction(songId restApiV1.SongId) {
	c.modified = true
	c.songIds = append(c.songIds, songId)
	c.RefreshView()
}

func (c *CurrentComponent) AddSongsFromAlbumAction(albumId restApiV1.AlbumId) {
	if albumId != restApiV1.UnknownAlbumId {
		for _, song := range c.app.localDb.AlbumOrderedSongs[albumId] {
			c.tryToAppendSong(song.Id)
		}
	} else {
		for _, song := range c.app.localDb.UnknownAlbumSongs {
			c.tryToAppendSong(song.Id)
		}
	}
	c.RefreshView()
}

func (c *CurrentComponent) AddSongsFromArtistAction(artistId restApiV1.ArtistId) {
	if artistId != restApiV1.UnknownArtistId {
		for _, song := range c.app.localDb.ArtistOrderedSongs[artistId] {
			c.tryToAppendSong(song.Id)
		}
	} else {
		for _, song := range c.app.localDb.UnknownArtistSongs {
			c.tryToAppendSong(song.Id)
		}
	}
	c.RefreshView()
}

func (c *CurrentComponent) AddSongsFromPlaylistAction(playlistId restApiV1.PlaylistId) {
	for _, songId := range c.app.localDb.Playlists[playlistId].SongIds {
		c.tryToAppendSong(songId)
	}
	c.RefreshView()
}

func (c *CurrentComponent) LoadSongsFromPlaylistAction(playlistId restApiV1.PlaylistId) {
	c.songIds = nil
	c.srcPlaylistId = &playlistId
	for _, songId := range c.app.localDb.Playlists[playlistId].SongIds {
		c.tryToAppendSong(songId)
	}
	c.modified = false
	c.RefreshView()
}

func (c *CurrentComponent) RemoveSongFromPlaylistAction(songIdx int) {
	c.songIds = append(c.songIds[0:songIdx], c.songIds[songIdx+1:]...)
	c.modified = true
	c.RefreshView()
}

func (c *CurrentComponent) AddSongsAction(songIds []restApiV1.SongId) {
	for _, songId := range songIds {
		c.tryToAppendSong(songId)
	}
	c.RefreshView()
}

func (c *CurrentComponent) tryToAppendSong(songId restApiV1.SongId) {
	// Don't append explicit songs if user profile ask for it
	if c.app.HideExplicitSongForConnectedUser() && c.app.localDb.Songs[songId].ExplicitFg {
		return
	}
	c.modified = true
	c.songIds = append(c.songIds, songId)
}
