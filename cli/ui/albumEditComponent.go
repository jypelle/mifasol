package ui

import (
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/rivo/tview"
)

type AlbumEditComponent struct {
	*tview.Form
	nameInputField  *tview.InputField
	uiApp           *UIApp
	albumId         string
	albumMeta       *restApiV1.AlbumMeta
	originPrimitive tview.Primitive
}

func OpenAlbumCreateComponent(uiApp *UIApp, originPrimitive tview.Primitive) {
	OpenAlbumEditComponent(uiApp, "", &restApiV1.AlbumMeta{}, originPrimitive)
}

func OpenAlbumEditComponent(uiApp *UIApp, albumId string, albumMeta *restApiV1.AlbumMeta, originPrimitive tview.Primitive) {

	// Only admin can create or edit an album
	if !uiApp.IsConnectedUserAdmin() {
		uiApp.WarningMessage("Only administrator can create or edit an album")
		return
	}

	c := &AlbumEditComponent{
		uiApp:           uiApp,
		albumId:         albumId,
		albumMeta:       albumMeta,
		originPrimitive: originPrimitive,
	}

	c.nameInputField = tview.NewInputField().
		SetLabel("Name").
		SetText(albumMeta.Name).
		SetFieldWidth(50)

	c.Form = tview.NewForm()
	c.Form.AddFormItem(c.nameInputField)
	c.Form.AddButton("Save", c.save)
	c.Form.AddButton("Cancel", c.cancel)
	if c.albumId != "" {
		c.Form.SetBorder(true).SetTitle("Edit album")
	} else {
		c.Form.SetBorder(true).SetTitle("Create album")
	}
	uiApp.pagesComponent.AddAndSwitchToPage("albumEdit", c, true)

}

func (c *AlbumEditComponent) save() {
	c.albumMeta.Name = c.nameInputField.GetText()
	if c.albumId != "" {
		_, cliErr := c.uiApp.restClient.UpdateAlbum(c.albumId, c.albumMeta)
		if cliErr != nil {
			c.uiApp.ClientErrorMessage("Unable to update the album", cliErr)
		}
	} else {
		_, cliErr := c.uiApp.restClient.CreateAlbum(c.albumMeta)
		if cliErr != nil {
			c.uiApp.ClientErrorMessage("Unable to create the album", cliErr)
		}
	}
	c.uiApp.Reload()

	c.close()
}

func (c *AlbumEditComponent) cancel() {
	c.close()
}

func (c *AlbumEditComponent) close() {
	c.uiApp.pagesComponent.RemovePage("albumEdit")
	c.uiApp.tviewApp.SetFocus(c.originPrimitive)
}
