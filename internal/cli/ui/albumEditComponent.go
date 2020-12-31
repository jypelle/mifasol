package ui

import (
	"github.com/jypelle/mifasol/restApiV1"
	"gitlab.com/tslocum/cview"
)

type AlbumEditComponent struct {
	*cview.Form
	nameInputField  *cview.InputField
	uiApp           *App
	albumId         restApiV1.AlbumId
	albumMeta       *restApiV1.AlbumMeta
	originPrimitive cview.Primitive
}

func OpenAlbumCreateComponent(uiApp *App, originPrimitive cview.Primitive) {
	OpenAlbumEditComponent(uiApp, "", &restApiV1.AlbumMeta{}, originPrimitive)
}

func OpenAlbumEditComponent(uiApp *App, albumId restApiV1.AlbumId, albumMeta *restApiV1.AlbumMeta, originPrimitive cview.Primitive) {

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

	c.nameInputField = cview.NewInputField()
	c.nameInputField.SetLabel("Name")
	c.nameInputField.SetText(albumMeta.Name)
	c.nameInputField.SetFieldWidth(50)

	c.Form = cview.NewForm()
	c.Form.SetFieldTextColorFocused(cview.Styles.PrimitiveBackgroundColor)
	c.Form.SetFieldBackgroundColorFocused(cview.Styles.PrimaryTextColor)

	c.Form.AddFormItem(c.nameInputField)
	c.Form.AddButton("Save", c.save)
	c.Form.AddButton("Cancel", c.cancel)
	if c.albumId != "" {
		c.Form.SetBorder(true)
		c.Form.SetTitle("Edit album")
	} else {
		c.Form.SetBorder(true)
		c.Form.SetTitle("Create album")
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
	c.uiApp.cviewApp.SetFocus(c.originPrimitive)
}
