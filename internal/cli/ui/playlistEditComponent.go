package ui

import (
	"code.rocketnine.space/tslocum/cview"
	"github.com/jypelle/mifasol/restApiV1"
	"strconv"
)

type PlaylistEditComponent struct {
	*cview.Form
	nameInputField  *cview.InputField
	ownerDropDowns  []*cview.DropDown
	uiApp           *App
	playlistId      restApiV1.PlaylistId
	playlistMeta    *restApiV1.PlaylistMeta
	originPrimitive cview.Primitive
}

func OpenPlaylistEditComponent(uiApp *App, playlistId restApiV1.PlaylistId, playlistMeta *restApiV1.PlaylistMeta, originPrimitive cview.Primitive) {

	// Only admin or playlist owner can edit a playlist
	if !uiApp.IsConnectedUserAdmin() && !uiApp.localDb.IsPlaylistOwnedBy(playlistId, uiApp.ConnectedUserId()) {
		uiApp.WarningMessage("Only administrator or playlist owner can edit a playlist")
		return
	}

	c := &PlaylistEditComponent{
		uiApp:           uiApp,
		playlistId:      playlistId,
		playlistMeta:    playlistMeta,
		originPrimitive: originPrimitive,
	}

	c.nameInputField = cview.NewInputField()
	c.nameInputField.SetLabel("Name")
	c.nameInputField.SetText(playlistMeta.Name)
	c.nameInputField.SetFieldWidth(50)

	c.Form = cview.NewForm()
	c.Form.SetFieldTextColorFocused(cview.Styles.PrimitiveBackgroundColor)
	c.Form.SetFieldBackgroundColorFocused(cview.Styles.PrimaryTextColor)

	c.Form.AddFormItem(c.nameInputField)
	for _, userId := range c.playlistMeta.OwnerUserIds {
		c.addOwner(userId)
	}
	c.addOwner("")

	c.Form.AddButton("Save", c.save)
	c.Form.AddButton("Cancel", c.cancel)
	c.Form.SetBorder(true)
	c.Form.SetTitle("Edit Playlist")

	uiApp.pagesComponent.AddAndSwitchToPage("playlistEdit", c, true)

}

func (c *PlaylistEditComponent) save() {
	// Name
	c.playlistMeta.Name = c.nameInputField.GetText()

	// Owners
	c.playlistMeta.OwnerUserIds = nil
	for _, ownerDropDown := range c.ownerDropDowns {
		selectedOwnerInd, _ := ownerDropDown.GetCurrentOption()
		var id restApiV1.UserId
		if selectedOwnerInd > 0 {
			id = c.uiApp.localDb.OrderedUsers[selectedOwnerInd-1].Id
			c.playlistMeta.OwnerUserIds = append(c.playlistMeta.OwnerUserIds, id)
		}
	}

	_, cliErr := c.uiApp.restClient.UpdatePlaylist(c.playlistId, c.playlistMeta)
	if cliErr != nil {
		c.uiApp.ClientErrorMessage("Unable to update the playlist", cliErr)
		return
	}

	c.close()
	c.uiApp.Reload()
}

func (c *PlaylistEditComponent) cancel() {
	c.close()
}

func (c *PlaylistEditComponent) addOwner(userId restApiV1.UserId) {
	ownerDropDown := cview.NewDropDown()
	ownerDropDown.SetLabel("Owner " + strconv.Itoa(len(c.ownerDropDowns)+1))
	selectedOwnerInd := 0
	ownerDropDown.AddOptionsSimple("(Nobody)")
	for ind, user := range c.uiApp.localDb.OrderedUsers {
		ownerDropDown.AddOptionsSimple(user.Name)
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
	c.uiApp.cviewApp.SetFocus(c.originPrimitive)
}
