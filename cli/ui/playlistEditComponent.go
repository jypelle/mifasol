package ui

import (
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/rivo/tview"
	"strconv"
)

type PlaylistEditComponent struct {
	*tview.Form
	nameInputField  *tview.InputField
	ownerDropDowns  []*tview.DropDown
	uiApp           *UIApp
	playlist        *restApiV1.Playlist
	originPrimitive tview.Primitive
}

func OpenPlaylistEditComponent(uiApp *UIApp, playlist *restApiV1.Playlist, originPrimitive tview.Primitive) {

	// Only admin or playlist owner can edit a playlist
	if !uiApp.IsConnectedUserAdmin() && !uiApp.localDb.IsPlaylistOwnedBy(playlist.Id, uiApp.ConnectedUserId()) {
		uiApp.WarningMessage("Only administrator or playlist owner can edit a playlist")
		return
	}

	c := &PlaylistEditComponent{
		uiApp:           uiApp,
		playlist:        playlist,
		originPrimitive: originPrimitive,
	}

	c.nameInputField = tview.NewInputField().
		SetLabel("Name").
		SetText(playlist.Name).
		SetFieldWidth(50)

	c.Form = tview.NewForm()
	c.Form.AddFormItem(c.nameInputField)
	for _, userId := range c.playlist.OwnerUserIds {
		c.addOwner(userId)
	}
	c.addOwner("")

	c.Form.AddButton("Save", c.save)
	c.Form.AddButton("Cancel", c.cancel)
	c.Form.SetBorder(true).SetTitle("Edit Playlist")

	uiApp.pagesComponent.AddAndSwitchToPage("playlistEdit", c, true)

}

func (c *PlaylistEditComponent) save() {
	// Name
	c.playlist.PlaylistMeta.Name = c.nameInputField.GetText()

	// Owners
	c.playlist.OwnerUserIds = nil
	for _, ownerDropDown := range c.ownerDropDowns {
		selectedOwnerInd, _ := ownerDropDown.GetCurrentOption()
		var id string
		if selectedOwnerInd > 0 {
			id = c.uiApp.localDb.OrderedUsers[selectedOwnerInd-1].Id
			c.playlist.OwnerUserIds = append(c.playlist.OwnerUserIds, id)
		}
	}

	_, cliErr := c.uiApp.restClient.UpdatePlaylist(c.playlist.Id, &c.playlist.PlaylistMeta)
	if cliErr != nil {
		c.uiApp.ClientErrorMessage("Unable to update the playlist", cliErr)
	}

	c.uiApp.Reload()

	c.close()
}

func (c *PlaylistEditComponent) cancel() {
	c.close()
}

func (c *PlaylistEditComponent) addOwner(userId string) {
	ownerDropDown := tview.NewDropDown().
		SetLabel("Owner " + strconv.Itoa(len(c.ownerDropDowns)+1))
	selectedOwnerInd := 0
	ownerDropDown.AddOption("(Nobody)", nil)
	for ind, user := range c.uiApp.localDb.OrderedUsers {
		ownerDropDown.AddOption(user.Name, nil)
		if userId == user.Id {
			selectedOwnerInd = ind + 1
		}
	}
	ownerDropDown.SetCurrentOption(selectedOwnerInd)
	c.ownerDropDowns = append(c.ownerDropDowns, ownerDropDown)
	c.Form.AddFormItem(ownerDropDown)
}

func (c *PlaylistEditComponent) close() {
	c.uiApp.pagesComponent.RemovePage("playlistEdit")
	c.uiApp.tviewApp.SetFocus(c.originPrimitive)
}
