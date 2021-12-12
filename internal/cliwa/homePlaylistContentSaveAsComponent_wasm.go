package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restApiV1"
	"html"
	"sort"
	"strings"
	"syscall/js"
)

type HomePlaylistContentSaveAsComponent struct {
	app              *App
	songIds          []restApiV1.SongId
	closed           bool
	newPlaylistName  string
	targetPlaylistId restApiV1.PlaylistId
}

func NewHomePlaylistContentSaveAsComponent(app *App, songIds []restApiV1.SongId) *HomePlaylistContentSaveAsComponent {
	c := &HomePlaylistContentSaveAsComponent{
		app:     app,
		songIds: songIds,
	}

	return c
}

func (c *HomePlaylistContentSaveAsComponent) Render() {
	div := jst.Id("homeMainModal")
	div.Set("innerHTML", c.app.RenderTemplate(
		nil, "home/playlistContentSaveAs/index"),
	)

	form := jst.Id("playlistContentSaveAsForm")
	form.Call("addEventListener", "submit", c.app.AddEventFuncPreventDefault(c.saveAction))
	cancelButton := jst.Id("playlistContentSaveAsCancelButton")
	cancelButton.Call("addEventListener", "click", c.app.AddEventFunc(c.cancelAction))

	// Playlist
	playlistCurrentBlock := jst.Id("playlistContentSaveAsPlaylistCurrentBlock")
	playlistCurrentName := jst.Id("playlistContentSaveAsPlaylistCurrentName")
	playlistCurrentDelete := jst.Id("playlistContentSaveAsPlaylistCurrentDelete")
	playlistSearchBlock := jst.Id("playlistContentSaveAsPlaylistSearchBlock")
	playlistSearchInput := jst.Id("playlistContentSaveAsPlaylistSearchInput")
	playlistSearchList := jst.Id("playlistContentSaveAsPlaylistSearchList")

	playlistCurrentDelete.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		c.targetPlaylistId = ""
		c.newPlaylistName = ""
		playlistCurrentName.Set("innerHTML", "")
		// Add searchInput
		playlistSearchBlock.Get("style").Set("display", "block")
		playlistCurrentBlock.Get("style").Set("display", "none")
	}))

	playlistSearchInput.Call("addEventListener", "keypress", c.app.AddBlockingRichEventFunc(func(this js.Value, i []js.Value) {
		if i[0].Get("which").Int() == 13 {
			i[0].Call("preventDefault")
		}
	}))
	playlistSearchInput.Call("addEventListener", "input", c.app.AddEventFunc(c.playlistSearchAction))
	playlistSearchInput.Call("addEventListener", "focusout", c.app.AddRichEventFunc(func(this js.Value, i []js.Value) {
		relatedTarget := i[0].Get("relatedTarget")
		if relatedTarget.Truthy() && relatedTarget.Call("closest", ".playlistLink, .newPlaylistLink").Truthy() {
			return
		}
		// Clear search input
		playlistSearchInput.Set("value", "")
		c.playlistSearchAction()
	}))
	playlistSearchList.Call("addEventListener", "click", c.app.AddRichEventFunc(func(this js.Value, i []js.Value) {
		link := i[0].Get("target").Call("closest", ".playlistLink, .newPlaylistLink")
		if !link.Truthy() {
			return
		}
		dataset := link.Get("dataset")

		switch link.Get("className").String() {
		case "playlistLink":
			playlistId := restApiV1.PlaylistId(dataset.Get("playlistid").String())
			c.targetPlaylistId = playlistId
			c.newPlaylistName = ""
			playlistCurrentName.Set("innerHTML", html.EscapeString(c.app.localDb.Playlists[playlistId].Name))

			// Clear
			playlistSearchInput.Set("value", "")
			c.playlistSearchAction()

			// Remove searchInput
			playlistSearchBlock.Get("style").Set("display", "none")
			playlistCurrentBlock.Get("style").Set("display", "flex")

		case "newPlaylistLink":
			c.targetPlaylistId = ""
			c.newPlaylistName = strings.TrimSpace(playlistSearchInput.Get("value").String())
			playlistCurrentName.Set("innerHTML", html.EscapeString(c.newPlaylistName))

			// Clear
			playlistSearchInput.Set("value", "")
			c.playlistSearchAction()

			// Remove searchInput
			playlistSearchBlock.Get("style").Set("display", "none")
			playlistCurrentBlock.Get("style").Set("display", "flex")
		}
	}))

}

func (c *HomePlaylistContentSaveAsComponent) saveAction() {
	if c.closed {
		return
	}

	if c.targetPlaylistId == "" && c.newPlaylistName == "" {
		c.app.HomeComponent.MessageComponent.WarningMessage("No playlist selected")
		return
	}

	c.app.ShowLoader("Updating playlist")
	defer c.app.HideLoader()

	var playlistMeta restApiV1.PlaylistMeta

	if c.targetPlaylistId != "" {
		playlistMeta = c.app.localDb.Playlists[c.targetPlaylistId].PlaylistMeta
		playlistMeta.SongIds = c.songIds

		_, cliErr := c.app.restClient.UpdatePlaylist(c.targetPlaylistId, &playlistMeta)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to update the playlist", cliErr)
			return
		}
		c.app.HomeComponent.CurrentComponent.srcPlaylistId = &c.targetPlaylistId
	} else {
		playlistMeta.Name = c.newPlaylistName
		playlistMeta.OwnerUserIds = append(playlistMeta.OwnerUserIds, c.app.ConnectedUserId())
		playlistMeta.SongIds = c.songIds

		newPlaylist, cliErr := c.app.restClient.CreatePlaylist(&playlistMeta)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to create the playlist", cliErr)
			return
		}
		c.app.HomeComponent.CurrentComponent.srcPlaylistId = &newPlaylist.Id
	}
	c.app.HomeComponent.CurrentComponent.modified = false
	c.close()
	c.app.HomeComponent.Reload()
}

func (c *HomePlaylistContentSaveAsComponent) cancelAction() {
	if c.closed {
		return
	}
	c.close()
}

func (c *HomePlaylistContentSaveAsComponent) close() {
	c.closed = true
	c.app.HomeComponent.CloseModal()
}

func (c *HomePlaylistContentSaveAsComponent) playlistSearchAction() {
	playlistSearchInput := jst.Id("playlistContentSaveAsPlaylistSearchInput")
	playlistSearchList := jst.Id("playlistContentSaveAsPlaylistSearchList")
	nameFilter := strings.TrimSpace(playlistSearchInput.Get("value").String())

	type PlaylistSearchItem struct {
		PlaylistId        restApiV1.PlaylistId
		PlaylistName      string
		PlaylistSongCount int
	}

	var resultPlaylistList []*PlaylistSearchItem

	if nameFilter != "" {
		lowerNameFilter := strings.ToLower(nameFilter)
		for _, playlist := range c.app.localDb.OrderedPlaylists {

			if
			// Only admin or playlist owner can update a playlist content
			(!c.app.localDb.IsPlaylistOwnedBy(playlist.Id, c.app.ConnectedUserId()) && !c.app.IsConnectedUserAdmin()) ||
				// Name filter should match
				!strings.Contains(strings.ToLower(playlist.Name), lowerNameFilter) {
				continue
			}

			playlistSearchItem := &PlaylistSearchItem{
				PlaylistId:        playlist.Id,
				PlaylistName:      playlist.Name,
				PlaylistSongCount: len(playlist.SongIds),
			}

			resultPlaylistList = append(resultPlaylistList, playlistSearchItem)
		}

		sort.SliceStable(resultPlaylistList, func(i, j int) bool {
			return len(resultPlaylistList[i].PlaylistName) < len(resultPlaylistList[j].PlaylistName)
		})

		if len(resultPlaylistList) > 100 {
			resultPlaylistList = resultPlaylistList[0:100]
		}

		playlistSearchList.Set("innerHTML", c.app.RenderTemplate(
			struct {
				PlaylistList []*PlaylistSearchItem
				NameFilter   string
			}{
				PlaylistList: resultPlaylistList,
				NameFilter:   nameFilter,
			},
			"home/playlistContentSaveAs/playlistSearchList"),
		)
		playlistSearchList.Get("style").Set("display", "block")
	} else {
		playlistSearchList.Set("innerHTML", "")
		playlistSearchList.Get("style").Set("display", "none")
	}

}
