package ui

import (
	"github.com/jypelle/mifasol/restApiV1"
	"gitlab.com/tslocum/cview"
)

type ArtistEditComponent struct {
	*cview.Form
	nameInputField  *cview.InputField
	uiApp           *App
	artistId        restApiV1.ArtistId
	artistMeta      *restApiV1.ArtistMeta
	originPrimitive cview.Primitive
}

func OpenArtistCreateComponent(uiApp *App, originPrimitive cview.Primitive) {
	OpenArtistEditComponent(uiApp, "", &restApiV1.ArtistMeta{}, originPrimitive)
}

func OpenArtistEditComponent(uiApp *App, artistId restApiV1.ArtistId, artistMeta *restApiV1.ArtistMeta, originPrimitive cview.Primitive) {

	// Only admin can create or edit an artist
	if !uiApp.IsConnectedUserAdmin() {
		uiApp.WarningMessage("Only administrator can create or edit an artist")
		return
	}

	c := &ArtistEditComponent{
		uiApp:           uiApp,
		artistId:        artistId,
		artistMeta:      artistMeta,
		originPrimitive: originPrimitive,
	}

	c.nameInputField = cview.NewInputField()
	c.nameInputField.SetLabel("Name")
	c.nameInputField.SetText(artistMeta.Name)
	c.nameInputField.SetFieldWidth(50)

	c.Form = cview.NewForm()
	c.Form.SetFieldTextColorFocused(cview.Styles.PrimitiveBackgroundColor)
	c.Form.SetFieldBackgroundColorFocused(cview.Styles.PrimaryTextColor)

	c.Form.AddFormItem(c.nameInputField)
	c.Form.AddButton("Save", c.save)
	c.Form.AddButton("Cancel", c.cancel)
	if c.artistId != "" {
		c.Form.SetBorder(true)
		c.Form.SetTitle("Edit artist")
	} else {
		c.Form.SetBorder(true)
		c.Form.SetTitle("Create artist")
	}

	uiApp.pagesComponent.AddAndSwitchToPage("artistEdit", c, true)
}

func (c *ArtistEditComponent) save() {
	c.artistMeta.Name = c.nameInputField.GetText()
	if c.artistId != "" {
		_, cliErr := c.uiApp.restClient.UpdateArtist(c.artistId, c.artistMeta)
		if cliErr != nil {
			c.uiApp.ClientErrorMessage("Unable to update the artist", cliErr)
		}
	} else {
		_, cliErr := c.uiApp.restClient.CreateArtist(c.artistMeta)
		if cliErr != nil {
			c.uiApp.ClientErrorMessage("Unable to create the artist", cliErr)
		}
	}
	c.uiApp.Reload()

	c.close()
}

func (c *ArtistEditComponent) cancel() {
	c.close()
}

func (c *ArtistEditComponent) close() {
	c.uiApp.pagesComponent.RemovePage("artistEdit")
	c.uiApp.cviewApp.SetFocus(c.originPrimitive)
}
