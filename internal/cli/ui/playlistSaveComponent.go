package ui

import (
	"github.com/jypelle/mifasol/restApiV1"
	"gitlab.com/tslocum/cview"
)

type PlaylistSaveComponent struct {
	*cview.Form
	playlistsDropDown        *cview.DropDown
	nameInputField           *cview.InputField
	uiApp                    *App
	srcPlaylistId            *restApiV1.PlaylistId
	songIds                  []restApiV1.SongId
	orderedFilteredPlaylists []*restApiV1.Playlist
	originPrimitive          cview.Primitive
}

func OpenPlaylistContentSaveComponent(uiApp *App, songIds []restApiV1.SongId, srcPlaylistId *restApiV1.PlaylistId, originPrimitive cview.Primitive) {

	// Only admin or playlist owner can edit playlist content
	if srcPlaylistId != nil && !uiApp.IsConnectedUserAdmin() && !uiApp.localDb.IsPlaylistOwnedBy(*srcPlaylistId, uiApp.ConnectedUserId()) {
		uiApp.WarningMessage("Only administrator or playlist owner can edit playlist content")
		return
	}

	c := &PlaylistSaveComponent{
		uiApp:           uiApp,
		srcPlaylistId:   srcPlaylistId,
		songIds:         songIds,
		originPrimitive: originPrimitive,
	}

	c.nameInputField = cview.NewInputField()
	c.nameInputField.SetLabel("Name")
	c.nameInputField.SetText("")
	c.nameInputField.SetFieldWidth(50)

	c.playlistsDropDown = cview.NewDropDown()
	c.playlistsDropDown.SetLabel("Playlist")
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
			c.playlistsDropDown.AddOptionsSimple("(New Playlist)")
		} else {
			c.playlistsDropDown.AddOptionsSimple(playList.Name)
			if srcPlaylistId != nil && *srcPlaylistId == playList.Id {
				selectedPlaylistInd = ind
			}
		}
	}
	c.playlistsDropDown.SetSelectedFunc(func(index int, option *cview.DropDownOption) {
		if index > 0 {
			c.nameInputField.SetText(c.orderedFilteredPlaylists[index].Name)
		} else {
			c.nameInputField.SetText("")
		}
	})
	c.playlistsDropDown.SetCurrentOption(selectedPlaylistInd)

	c.Form = cview.NewForm()
	c.Form.SetFieldTextColorFocused(cview.Styles.PrimitiveBackgroundColor)
	c.Form.SetFieldBackgroundColorFocused(cview.Styles.PrimaryTextColor)

	c.Form.AddFormItem(c.playlistsDropDown)
	c.Form.AddFormItem(c.nameInputField)
	c.Form.AddButton("Save", c.save)
	c.Form.AddButton("Cancel", c.cancel)
	c.Form.SetBorder(true)
	c.Form.SetTitle("Save playlist content")

	uiApp.pagesComponent.AddAndSwitchToPage("playlistContentSave", c, true)

}

func (c *PlaylistSaveComponent) save() {
	selectedPlaylistInd, _ := c.playlistsDropDown.GetCurrentOption()
	var id restApiV1.PlaylistId
	var playlistMeta restApiV1.PlaylistMeta

	if selectedPlaylistInd == 0 {
		playlistMeta.Name = c.nameInputField.GetText()
		playlistMeta.SongIds = c.songIds
		playlistMeta.OwnerUserIds = append(playlistMeta.OwnerUserIds, c.uiApp.ConnectedUserId())

		playList, cliErr := c.uiApp.restClient.CreatePlaylist(&playlistMeta)
		if cliErr != nil {
			c.uiApp.ClientErrorMessage("Unable to create the playlist", cliErr)
			return
		}

		id = playList.Id
	} else {
		selectedPlaylist := c.orderedFilteredPlaylists[selectedPlaylistInd]
		playlistMeta = selectedPlaylist.PlaylistMeta
		playlistMeta.Name = c.nameInputField.GetText()
		playlistMeta.SongIds = c.songIds

		_, cliErr := c.uiApp.restClient.UpdatePlaylist(selectedPlaylist.Id, &playlistMeta)
		if cliErr != nil {
			c.uiApp.ClientErrorMessage("Unable to create the playlist", cliErr)
			return
		}

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
	c.uiApp.cviewApp.SetFocus(c.originPrimitive)
}
