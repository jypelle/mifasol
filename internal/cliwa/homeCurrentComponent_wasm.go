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

type HomeCurrentComponent struct {
	app *App

	songIds        []restApiV1.SongId
	currentSongIdx int
	srcPlaylistId  *restApiV1.PlaylistId
	modified       bool
}

func NewHomeCurrentComponent(app *App) *HomeCurrentComponent {
	c := &HomeCurrentComponent{
		app:            app,
		modified:       true,
		currentSongIdx: -1,
	}

	return c
}

func (c *HomeCurrentComponent) Show() {
	div := jst.Id("currentComponent")
	div.Set("innerHTML", c.app.RenderTemplate(
		nil, "home/current/index"),
	)

	currentCleanButton := jst.Id("currentCleanButton")
	currentCleanButton.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		c.songIds = nil
		c.srcPlaylistId = nil
		c.modified = true
		c.currentSongIdx = -1
		c.RefreshView()
	}))
	currentShuffleButton := jst.Id("currentShuffleButton")
	currentShuffleButton.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		rand.Shuffle(len(c.songIds), func(i, j int) { c.songIds[i], c.songIds[j] = c.songIds[j], c.songIds[i] })
		c.modified = true
		c.currentSongIdx = -1
		c.RefreshView()
	}))
	currentSaveButton := jst.Id("currentSaveButton")
	currentSaveButton.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		if c.modified {
			if c.srcPlaylistId == nil {
				// Save content as new playlist
				component := NewHomePlaylistContentSaveAsComponent(c.app, c.songIds)

				c.app.HomeComponent.OpenModal()
				component.Show()
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
	currentSaveAsButton := jst.Id("currentSaveAsButton")
	currentSaveAsButton.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		// Save content as
		component := NewHomePlaylistContentSaveAsComponent(c.app, c.songIds)
		c.app.HomeComponent.OpenModal()
		component.Show()
	}))

	listDiv := jst.Id("currentList")
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

func (c *HomeCurrentComponent) PlayNextSongAction() {
	if c.currentSongIdx != -1 && c.currentSongIdx < len(c.songIds)-1 {
		c.currentSongIdx++
		c.app.HomeComponent.PlayerComponent.PlaySongAction(c.songIds[c.currentSongIdx])
	} else {
		c.currentSongIdx = -1
	}
}

func (c *HomeCurrentComponent) RefreshView() {
	oldSongIds := c.songIds

	c.songIds = []restApiV1.SongId{}

	// Remove deleted songId
	for _, songId := range oldSongIds {
		if _, ok := c.app.localDb.Songs[songId]; ok {
			c.songIds = append(c.songIds, songId)
		} else {
			if len(c.songIds) < c.currentSongIdx {
				c.currentSongIdx--
			} else if len(c.songIds) == c.currentSongIdx {
				c.currentSongIdx = -1
			}
		}
	}

	if c.srcPlaylistId != nil {
		if _, ok := c.app.localDb.Playlists[*c.srcPlaylistId]; !ok {
			// If src playlist has been deleted, current playlist is a new playlist
			c.modified = true
			c.srcPlaylistId = nil
		}
	}

	// Update current playlist title
	titleSpan := jst.Id("currentTitle")
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

	// Update list
	var divContent strings.Builder
	for songIdx, songId := range c.songIds {
		divContent.WriteString(c.addSongItem(songIdx, c.app.localDb.Songs[songId]))
	}
	listDiv := jst.Id("currentList")
	listDiv.Set("innerHTML", divContent.String())
}

func (c *HomeCurrentComponent) addSongItem(songIdx int, song *restApiV1.Song) string {
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
		IsPlaying bool
	}{
		SongId:    string(song.Id),
		SongIdx:   songIdx,
		SongName:  song.Name,
		IsPlaying: songIdx == c.currentSongIdx,
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

	divContent.WriteString(c.app.RenderTemplate(&songItem, "home/current/songItem"))

	return divContent.String()
}

func (c *HomeCurrentComponent) AddSongAction(songId restApiV1.SongId) {
	c.modified = true
	c.songIds = append(c.songIds, songId)
	c.RefreshView()
}

func (c *HomeCurrentComponent) AddSongsFromAlbumAction(albumId restApiV1.AlbumId) {
	if albumId != restApiV1.UnknownAlbumId {
		for _, song := range c.app.localDb.AlbumOrderedSongs[albumId] {
			c.tryToAppendSong(song)
		}
	} else {
		for _, song := range c.app.localDb.UnknownAlbumSongs {
			c.tryToAppendSong(song)
		}
	}
	c.RefreshView()
}

func (c *HomeCurrentComponent) AddSongsFromArtistAction(artistId restApiV1.ArtistId) {
	if artistId != restApiV1.UnknownArtistId {
		for _, song := range c.app.localDb.ArtistOrderedSongs[artistId] {
			c.tryToAppendSong(song)
		}
	} else {
		for _, song := range c.app.localDb.UnknownArtistSongs {
			c.tryToAppendSong(song)
		}
	}
	c.RefreshView()
}

func (c *HomeCurrentComponent) AddSongsFromPlaylistAction(playlistId restApiV1.PlaylistId) {
	for _, songId := range c.app.localDb.Playlists[playlistId].SongIds {
		c.tryToAppendSong(c.app.localDb.Songs[songId])
	}
	c.RefreshView()
}

func (c *HomeCurrentComponent) LoadSongsFromPlaylistAction(playlistId restApiV1.PlaylistId) {
	c.songIds = nil
	c.srcPlaylistId = &playlistId
	for _, songId := range c.app.localDb.Playlists[playlistId].SongIds {
		c.tryToAppendSong(c.app.localDb.Songs[songId])
	}
	c.modified = false
	c.currentSongIdx = -1
	c.RefreshView()
}

func (c *HomeCurrentComponent) RemoveSongFromPlaylistAction(songIdx int) {
	if songIdx < c.currentSongIdx {
		c.currentSongIdx--
	} else if songIdx == c.currentSongIdx {
		c.currentSongIdx = -1
	}

	c.songIds = append(c.songIds[0:songIdx], c.songIds[songIdx+1:]...)
	c.modified = true
	c.RefreshView()
}

func (c *HomeCurrentComponent) AddSongsAction(songs []*restApiV1.Song) {
	for _, song := range songs {
		c.tryToAppendSong(song)
	}
	c.RefreshView()
}

func (c *HomeCurrentComponent) tryToAppendSong(song *restApiV1.Song) {
	// Don't append explicit songs if user profile ask for it
	if c.app.HideExplicitSongForConnectedUser() && song.ExplicitFg {
		return
	}
	c.modified = true
	c.songIds = append(c.songIds, song.Id)
}
