package cliwa

import (
	"fmt"
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/sirupsen/logrus"
	"html"
	"math/rand"
	"strconv"
	"syscall/js"
)

type HomeCurrentComponent struct {
	app *App

	songIds        []restApiV1.SongId
	currentSongIdx int
	srcPlaylistId  *restApiV1.PlaylistId
	modified       bool

	displayedPage int
}

func NewHomeCurrentComponent(app *App) *HomeCurrentComponent {
	c := &HomeCurrentComponent{
		app:            app,
		modified:       true,
		currentSongIdx: -1,
	}

	return c
}

func (c *HomeCurrentComponent) Render() {
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
		c.displayedPage = 0
		c.RefreshView(0, true)
	}))
	currentShuffleButton := jst.Id("currentShuffleButton")
	currentShuffleButton.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		rand.Shuffle(len(c.songIds), func(i, j int) { c.songIds[i], c.songIds[j] = c.songIds[j], c.songIds[i] })
		c.modified = true
		c.currentSongIdx = -1
		c.displayedPage = 0
		c.RefreshView(0, true)
	}))
	currentSaveButton := jst.Id("currentSaveButton")
	currentSaveButton.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		if c.modified {
			if c.srcPlaylistId == nil {
				// Save content as new playlist
				component := NewHomePlaylistContentSaveAsComponent(c.app, c.songIds)

				c.app.HomeComponent.OpenModal()
				component.Render()
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
		component.Render()
	}))

	listDiv := jst.Id("currentList")
	listDiv.Call("addEventListener", "scroll", c.app.AddRichEventFunc(func(this js.Value, i []js.Value) {
		scrollHeight := listDiv.Get("scrollHeight").Int()
		scrollTop := listDiv.Get("scrollTop").Int()
		clientHeight := listDiv.Get("clientHeight").Int()
		if scrollTop+clientHeight >= scrollHeight-5 {
			if LibraryPageSize*(c.displayedPage+2) <= len(c.songIds) {
				c.displayedPage++
				logrus.Infof("scroll: Down %d / %d / %d / %d", c.displayedPage, scrollHeight, scrollTop+clientHeight, clientHeight)
				c.RefreshView(1, false)
			}
		} else if scrollTop == 0 {
			if c.displayedPage > 0 {
				c.displayedPage--
				logrus.Infof("scroll: Up %d / %d / %d / %d", c.displayedPage, scrollHeight, scrollTop+clientHeight, clientHeight)
				c.RefreshView(-1, false)
			}
		}
	}))
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
			c.RefreshView(0, false)
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
	c.RefreshView(0, false)
}

func (c *HomeCurrentComponent) RefreshView(direction int, resetPosition bool) {
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

	listDiv := jst.Id("currentList")

	if resetPosition {
		listDiv.Set("scrollTop", 0)
	}

	var divContentPreviousPage string
	var divContentCurrentPage string
	var divContentNextPage string

	// Refresh library list
	minIdx := LibraryPageSize * (c.displayedPage - 1)
	if minIdx < 0 {
		minIdx = 0
	}
	maxIdx := LibraryPageSize * (c.displayedPage + 2)
	if maxIdx > len(c.songIds) {
		maxIdx = len(c.songIds)
	}

	step1Idx := LibraryPageSize * c.displayedPage
	if step1Idx > maxIdx {
		step1Idx = maxIdx
	}
	step2Idx := LibraryPageSize * (c.displayedPage + 1)
	if step2Idx > maxIdx {
		step2Idx = maxIdx
	}

	divContentPreviousPage = c.renderSongItemList(c.songIds, minIdx, step1Idx)
	divContentCurrentPage = c.renderSongItemList(c.songIds, step1Idx, step2Idx)
	divContentNextPage = c.renderSongItemList(c.songIds, step2Idx, maxIdx)

	var newScrollTop int
	listDiv.Set("innerHTML", divContentPreviousPage)
	if direction == -1 {
		newScrollTop = listDiv.Get("scrollHeight").Int()
	}
	listDiv.Call("insertAdjacentHTML", "beforeEnd", divContentCurrentPage)
	if direction == 1 {
		newScrollTop = listDiv.Get("scrollHeight").Int() - listDiv.Get("clientHeight").Int()
	}
	listDiv.Call("insertAdjacentHTML", "beforeEnd", divContentNextPage)

	if direction != 0 {
		listDiv.Set("scrollTop", newScrollTop)
	}
}

func (c *HomeCurrentComponent) AddSongAction(songId restApiV1.SongId) {
	c.modified = true
	c.songIds = append(c.songIds, songId)
	c.RefreshView(0, false)
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
	c.RefreshView(0, false)
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
	c.RefreshView(0, false)
}

func (c *HomeCurrentComponent) AddSongsFromPlaylistAction(playlistId restApiV1.PlaylistId) {
	for _, songId := range c.app.localDb.Playlists[playlistId].SongIds {
		c.tryToAppendSong(c.app.localDb.Songs[songId])
	}
	c.RefreshView(0, false)
}

func (c *HomeCurrentComponent) LoadSongsFromPlaylistAction(playlistId restApiV1.PlaylistId) {
	c.songIds = nil
	c.srcPlaylistId = &playlistId
	for _, songId := range c.app.localDb.Playlists[playlistId].SongIds {
		c.tryToAppendSong(c.app.localDb.Songs[songId])
	}
	c.modified = false
	c.currentSongIdx = -1
	c.displayedPage = 0
	c.RefreshView(0, true)
}

func (c *HomeCurrentComponent) RemoveSongFromPlaylistAction(songIdx int) {
	if songIdx < c.currentSongIdx {
		c.currentSongIdx--
	} else if songIdx == c.currentSongIdx {
		c.currentSongIdx = -1
	}

	c.songIds = append(c.songIds[0:songIdx], c.songIds[songIdx+1:]...)
	c.modified = true
	c.RefreshView(0, false)
}

func (c *HomeCurrentComponent) RemoveDeletedSongsOrPlaylist() {
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

	c.displayedPage = 0
	c.RefreshView(0, true)
}

func (c *HomeCurrentComponent) AddSongsAction(songs []*restApiV1.Song) {
	for _, song := range songs {
		c.tryToAppendSong(song)
	}
	c.RefreshView(0, false)
}

func (c *HomeCurrentComponent) tryToAppendSong(song *restApiV1.Song) {
	// Don't append explicit songs if user profile ask for it
	if c.app.HideExplicitSongForConnectedUser() && song.ExplicitFg {
		return
	}
	c.modified = true
	c.songIds = append(c.songIds, song.Id)
}

func (c *HomeCurrentComponent) renderSongItemList(songIdList []restApiV1.SongId, minIdx, maxIdx int) string {

	// Update list
	type SongItem struct {
		SongId    string
		SongIdx   int
		SongName  string
		AlbumId   *string
		AlbumName string
		Artists   []struct {
			ArtistId   string
			ArtistName string
		}
		ExplicitFg bool
		IsPlaying  bool
	}

	var songItemList = make([]SongItem, maxIdx-minIdx)

	var song *restApiV1.Song

	for idx, songId := range songIdList[minIdx:maxIdx] {
		song = c.app.localDb.Songs[songId]
		songItemList[idx].SongId = string(song.Id)
		songItemList[idx].SongIdx = minIdx + idx
		songItemList[idx].SongName = song.Name
		songItemList[idx].IsPlaying = songItemList[idx].SongIdx == c.currentSongIdx

		if song.AlbumId != restApiV1.UnknownAlbumId {
			songItemList[idx].AlbumName = c.app.localDb.Albums[song.AlbumId].Name
			songItemList[idx].AlbumId = (*string)(&song.AlbumId)
			songItemList[idx].ExplicitFg = song.ExplicitFg
		}

		for _, artistId := range song.ArtistIds {
			songItemList[idx].Artists = append(songItemList[idx].Artists, struct {
				ArtistId   string
				ArtistName string
			}{
				ArtistId:   string(artistId),
				ArtistName: c.app.localDb.Artists[artistId].Name,
			})
		}
	}

	return c.app.RenderTemplate(songItemList, "home/current/songItemList")
}
