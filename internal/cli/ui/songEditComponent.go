package ui

import (
	"codeberg.org/tslocum/cview"
	"github.com/jypelle/mifasol/restApiV1"
	"strconv"
)

type SongEditComponent struct {
	*cview.Form
	nameInputField            *cview.InputField
	publicationYearInputField *cview.InputField
	albumDropDown             *cview.DropDown
	trackNumberInputField     *cview.InputField
	explicitFgCheckbox        *cview.CheckBox
	artistDropDowns           []*cview.DropDown
	uiApp                     *App
	song                      *restApiV1.Song
	originPrimitive           cview.Primitive
}

func OpenSongEditComponent(uiApp *App, song *restApiV1.Song, originPrimitive cview.Primitive) {

	// Only admin can edit a song
	if !uiApp.IsConnectedUserAdmin() {
		uiApp.WarningMessage("Only administrator can edit a song")
		return
	}

	c := &SongEditComponent{
		uiApp:           uiApp,
		song:            song,
		originPrimitive: originPrimitive,
	}

	// Name
	c.nameInputField = cview.NewInputField()
	c.nameInputField.SetLabel("Name")
	c.nameInputField.SetText(song.Name)
	c.nameInputField.SetFieldWidth(50)

	// Publication year
	c.publicationYearInputField = cview.NewInputField()
	c.publicationYearInputField.SetLabel("Publication year")
	c.publicationYearInputField.SetFieldWidth(4)

	if song.PublicationYear != nil {
		c.publicationYearInputField.SetText(strconv.FormatInt(*song.PublicationYear, 10))
	}

	// Album
	c.albumDropDown = cview.NewDropDown()
	c.albumDropDown.SetLabel("Album")
	selectedAlbumInd := 0
	for ind, album := range uiApp.localDb.OrderedAlbums {
		if ind == 0 {
			c.albumDropDown.AddOptionsSimple("(Unknown album)")
		} else {
			c.albumDropDown.AddOptionsSimple(album.Name)
			if song.AlbumId != restApiV1.UnknownAlbumId && song.AlbumId == album.Id {
				selectedAlbumInd = ind
			}
		}
	}
	c.albumDropDown.SetCurrentOption(selectedAlbumInd)

	// Track number
	c.trackNumberInputField = cview.NewInputField()
	c.trackNumberInputField.SetLabel("Track number")
	c.trackNumberInputField.SetFieldWidth(4)

	if song.TrackNumber != nil {
		c.trackNumberInputField.SetText(strconv.FormatInt(*song.TrackNumber, 10))
	}

	// Explicit flag
	c.explicitFgCheckbox = cview.NewCheckBox()
	c.explicitFgCheckbox.SetLabel("Explicit")
	c.explicitFgCheckbox.SetChecked(song.ExplicitFg)

	c.Form = cview.NewForm()
	c.Form.SetFieldTextColorFocused(cview.Styles.PrimitiveBackgroundColor)
	c.Form.SetFieldBackgroundColorFocused(cview.Styles.PrimaryTextColor)

	c.Form.AddFormItem(c.nameInputField)
	c.Form.AddFormItem(c.publicationYearInputField)
	c.Form.AddFormItem(c.albumDropDown)
	c.Form.AddFormItem(c.trackNumberInputField)
	c.Form.AddFormItem(c.explicitFgCheckbox)

	for _, artistId := range c.song.ArtistIds {
		c.addArtist(artistId)
	}
	c.addArtist("")

	c.Form.AddButton("Save", c.save)
	c.Form.AddButton("Cancel", c.cancel)
	c.Form.SetBorder(true)
	c.Form.SetTitle("Edit Song")
	uiApp.pagesComponent.AddAndSwitchToPage("songEdit", c, true)

}

func (c *SongEditComponent) save() {

	// Name
	c.song.SongMeta.Name = c.nameInputField.GetText()

	// Publication year
	c.song.SongMeta.PublicationYear = nil
	if c.publicationYearInputField.GetText() != "" {

		publicationYear, err := strconv.ParseInt(c.publicationYearInputField.GetText(), 10, 64)
		if err == nil {
			c.song.SongMeta.PublicationYear = &publicationYear
		}
	}

	// Album
	selectedAlbumInd, _ := c.albumDropDown.GetCurrentOption()
	var id restApiV1.AlbumId
	if selectedAlbumInd == 0 {
		c.song.SongMeta.AlbumId = restApiV1.UnknownAlbumId
	} else {
		id = c.uiApp.localDb.OrderedAlbums[selectedAlbumInd].Id
		c.song.SongMeta.AlbumId = id
	}

	// Artists
	c.song.ArtistIds = nil
	for _, artistDropDown := range c.artistDropDowns {
		selectedArtistInd, _ := artistDropDown.GetCurrentOption()
		var id restApiV1.ArtistId
		if selectedArtistInd > 0 {
			id = c.uiApp.localDb.OrderedArtists[selectedArtistInd].Id
			c.song.ArtistIds = append(c.song.ArtistIds, id)
		}
	}

	// Track number
	c.song.SongMeta.TrackNumber = nil
	if c.trackNumberInputField.GetText() != "" {

		trackNumber, err := strconv.ParseInt(c.trackNumberInputField.GetText(), 10, 64)
		if err == nil {
			c.song.SongMeta.TrackNumber = &trackNumber
		}
	}

	// Explicit flag
	c.song.SongMeta.ExplicitFg = c.explicitFgCheckbox.IsChecked()

	_, cliErr := c.uiApp.restClient.UpdateSong(c.song.Id, &c.song.SongMeta)
	if cliErr != nil {
		c.uiApp.ClientErrorMessage("Unable to update the song", cliErr)
		return
	}

	c.close()
	c.uiApp.Reload()
}

func (c *SongEditComponent) cancel() {
	c.close()
}

func (c *SongEditComponent) addArtist(artistId restApiV1.ArtistId) {
	artistDropDown := cview.NewDropDown()
	artistDropDown.SetLabel("Artist " + strconv.Itoa(len(c.artistDropDowns)+1))
	selectedArtistInd := 0
	for ind, artist := range c.uiApp.localDb.OrderedArtists {
		if ind == 0 {
			artistDropDown.AddOptionsSimple("(Unknown artist)")
		} else {
			artistDropDown.AddOptionsSimple(artist.Name)
			if artistId == artist.Id {
				selectedArtistInd = ind
			}
		}
	}
	artistDropDown.SetCurrentOption(selectedArtistInd)
	c.artistDropDowns = append(c.artistDropDowns, artistDropDown)
	c.Form.AddFormItem(artistDropDown)
}

func (c *SongEditComponent) close() {
	c.uiApp.pagesComponent.RemovePage("songEdit")
	c.uiApp.cviewApp.SetFocus(c.originPrimitive)
}
