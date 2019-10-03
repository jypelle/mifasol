package ui

import (
	"github.com/rivo/tview"
	"lyra/restApiV1"
)

type PlaylistSaveComponent struct {
	*tview.Form
	playlistsDropDown        *tview.DropDown
	nameInputField           *tview.InputField
	uiApp                    *UIApp
	srcPlaylistId            *string
	songIds                  []string
	orderedFilteredPlaylists []*restApiV1.Playlist
	originPrimitive          tview.Primitive
}

func OpenPlaylistContentSaveComponent(uiApp *UIApp, songIds []string, srcPlaylistId *string, originPrimitive tview.Primitive) {

	c := &PlaylistSaveComponent{
		uiApp:           uiApp,
		srcPlaylistId:   srcPlaylistId,
		songIds:         songIds,
		originPrimitive: originPrimitive,
	}

	c.nameInputField = tview.NewInputField().
		SetLabel("Name").
		SetText("").
		SetFieldWidth(50)

	c.playlistsDropDown = tview.NewDropDown().SetLabel("Playlist")
	c.orderedFilteredPlaylists = append(c.orderedFilteredPlaylists, nil)
	selectedPlaylistInd := 0
	for _, playlist := range uiApp.localDb.OrderedPlaylists {
		// Only admin or playlist owner can update a playlist content
		if uiApp.localDb.IsPlaylistOwnedBy(playlist.Id, uiApp.ConnectedUserId()) || uiApp.IsConnectedUserAdmin() {
			c.orderedFilteredPlaylists = append(c.orderedFilteredPlaylists, playlist)
		}
	}
	for ind, playList := range c.orderedFilteredPlaylists {
		if ind == 0 {
			c.playlistsDropDown.AddOption("(New Playlist)", nil)
		} else {
			c.playlistsDropDown.AddOption(playList.Name, nil)
			if srcPlaylistId != nil && *srcPlaylistId == playList.Id {
				selectedPlaylistInd = ind
			}
		}
	}
	c.playlistsDropDown.SetSelectedFunc(func(text string, index int) {
		if index > 0 {
			c.nameInputField.SetText(c.orderedFilteredPlaylists[index].Name)
		} else {
			c.nameInputField.SetText("")
		}
	})
	c.playlistsDropDown.SetCurrentOption(selectedPlaylistInd)

	c.Form = tview.NewForm()
	c.Form.AddFormItem(c.playlistsDropDown)
	c.Form.AddFormItem(c.nameInputField)
	c.Form.AddButton("Save", c.save)
	c.Form.AddButton("Cancel", c.cancel)
	c.Form.SetBorder(true).SetTitle("Save playlist content")

	uiApp.pagesComponent.AddAndSwitchToPage("playlistContentSave", c, true)

}

func (c *PlaylistSaveComponent) save() {
	selectedPlaylistInd, _ := c.playlistsDropDown.GetCurrentOption()
	var id string
	var playlistMeta restApiV1.PlaylistMeta

	if selectedPlaylistInd == 0 {
		playlistMeta.Name = c.nameInputField.GetText()
		playlistMeta.SongIds = c.songIds
		playlistMeta.OwnerUserIds = append(playlistMeta.OwnerUserIds, c.uiApp.ConnectedUserId())

		playList, _ := c.uiApp.restClient.CreatePlaylist(&playlistMeta)
		id = playList.Id
	} else {
		selectedPlaylist := c.orderedFilteredPlaylists[selectedPlaylistInd]
		playlistMeta = selectedPlaylist.PlaylistMeta
		playlistMeta.Name = c.nameInputField.GetText()
		playlistMeta.SongIds = c.songIds

		c.uiApp.restClient.UpdatePlaylist(selectedPlaylist.Id, &playlistMeta)
		id = selectedPlaylist.Id
	}
	c.uiApp.currentComponent.srcPlaylistId = &id
	c.uiApp.Reload()
	c.uiApp.currentComponent.SetModified(false)
	c.close()
}

func (c *PlaylistSaveComponent) cancel() {
	c.close()
}

func (c *PlaylistSaveComponent) close() {
	c.uiApp.pagesComponent.RemovePage("playlistContentSave")
	c.uiApp.tviewApp.SetFocus(c.originPrimitive)
}
