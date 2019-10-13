package ui

import (
	"github.com/rivo/tview"
	"lyra/restApiV1"
	"strconv"
)

type SongEditComponent struct {
	*tview.Form
	nameInputField            *tview.InputField
	publicationYearInputField *tview.InputField
	albumDropDown             *tview.DropDown
	trackNumberInputField     *tview.InputField
	artistDropDowns           []*tview.DropDown
	uiApp                     *UIApp
	song                      *restApiV1.Song
	originPrimitive           tview.Primitive
}

func OpenSongEditComponent(uiApp *UIApp, song *restApiV1.Song, originPrimitive tview.Primitive) {

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
	c.nameInputField = tview.NewInputField().
		SetLabel("Name").
		SetText(song.Name).
		SetFieldWidth(50)

	// Publication year
	c.publicationYearInputField = tview.NewInputField().
		SetLabel("Publication year").
		SetFieldWidth(4)

	if song.PublicationYear != nil {
		c.publicationYearInputField.SetText(strconv.FormatInt(*song.PublicationYear, 10))
	}

	// Album
	c.albumDropDown = tview.NewDropDown().
		SetLabel("Album")
	selectedAlbumInd := 0
	for ind, album := range uiApp.localDb.OrderedAlbums {
		if ind == 0 {
			c.albumDropDown.AddOption("(Unknown album)", nil)
		} else {
			c.albumDropDown.AddOption(album.Name, nil)
			if song.AlbumId != "" && song.AlbumId == album.Id {
				selectedAlbumInd = ind
			}
		}
	}
	c.albumDropDown.SetCurrentOption(selectedAlbumInd)

	// Track number
	c.trackNumberInputField = tview.NewInputField().
		SetLabel("Track number").
		SetFieldWidth(4)

	if song.TrackNumber != nil {
		c.trackNumberInputField.SetText(strconv.FormatInt(*song.TrackNumber, 10))
	}

	c.Form = tview.NewForm()
	c.Form.AddFormItem(c.nameInputField)
	c.Form.AddFormItem(c.publicationYearInputField)
	c.Form.AddFormItem(c.albumDropDown)
	c.Form.AddFormItem(c.trackNumberInputField)

	for _, artistId := range c.song.ArtistIds {
		c.addArtist(artistId)
	}
	c.addArtist("")

	c.Form.AddButton("Save", c.save)
	c.Form.AddButton("Cancel", c.cancel)
	c.Form.SetBorder(true).SetTitle("Edit Song")
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
	var id string
	if selectedAlbumInd == 0 {
		c.song.SongMeta.AlbumId = ""
	} else {
		id = c.uiApp.localDb.OrderedAlbums[selectedAlbumInd].Id
		c.song.SongMeta.AlbumId = id
	}

	// Artists
	c.song.ArtistIds = nil
	for _, artistDropDown := range c.artistDropDowns {
		selectedArtistInd, _ := artistDropDown.GetCurrentOption()
		var id string
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

	c.uiApp.restClient.UpdateSong(c.song.Id, &c.song.SongMeta)
	c.uiApp.Reload()

	c.close()
}

func (c *SongEditComponent) cancel() {
	c.close()
}

func (c *SongEditComponent) addArtist(artistId string) {
	artistDropDown := tview.NewDropDown().
		SetLabel("Artist " + strconv.Itoa(len(c.artistDropDowns)+1))
	selectedArtistInd := 0
	for ind, artist := range c.uiApp.localDb.OrderedArtists {
		if ind == 0 {
			artistDropDown.AddOption("(Unknown artist)", nil)
		} else {
			artistDropDown.AddOption(artist.Name, nil)
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
	c.uiApp.tviewApp.SetFocus(c.originPrimitive)
}
