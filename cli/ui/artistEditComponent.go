package ui

import (
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/rivo/tview"
)

type ArtistEditComponent struct {
	*tview.Form
	nameInputField  *tview.InputField
	uiApp           *UIApp
	artistId        restApiV1.ArtistId
	artistMeta      *restApiV1.ArtistMeta
	originPrimitive tview.Primitive
}

func OpenArtistCreateComponent(uiApp *UIApp, originPrimitive tview.Primitive) {
	OpenArtistEditComponent(uiApp, "", &restApiV1.ArtistMeta{}, originPrimitive)
}

func OpenArtistEditComponent(uiApp *UIApp, artistId restApiV1.ArtistId, artistMeta *restApiV1.ArtistMeta, originPrimitive tview.Primitive) {

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

	c.nameInputField = tview.NewInputField().
		SetLabel("Name").
		SetText(artistMeta.Name).
		SetFieldWidth(50)

	c.Form = tview.NewForm()
	c.Form.AddFormItem(c.nameInputField)
	c.Form.AddButton("Save", c.save)
	c.Form.AddButton("Cancel", c.cancel)
	if c.artistId != "" {
		c.Form.SetBorder(true).SetTitle("Edit artist")
	} else {
		c.Form.SetBorder(true).SetTitle("Create artist")
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
	c.uiApp.tviewApp.SetFocus(c.originPrimitive)
}
