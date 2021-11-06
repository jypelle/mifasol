package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restApiV1"
	"sort"
	"strconv"
	"strings"
	"syscall/js"
)

type HomeSongEditComponent struct {
	app      *App
	songId   restApiV1.SongId
	songMeta *restApiV1.SongMeta
	closed   bool
}

func NewHomeSongEditComponent(app *App, songId restApiV1.SongId, songMeta *restApiV1.SongMeta) *HomeSongEditComponent {
	c := &HomeSongEditComponent{
		app:      app,
		songId:   songId,
		songMeta: songMeta.Copy(),
	}

	return c
}

func (c *HomeSongEditComponent) Show() {
	songItem := struct {
		SongMeta  *restApiV1.SongMeta
		AlbumName string
	}{
		SongMeta: c.songMeta,
	}

	if c.songMeta.AlbumId != restApiV1.UnknownAlbumId {
		songItem.AlbumName = c.app.localDb.Albums[c.songMeta.AlbumId].Name
	}

	div := jst.Document.Call("getElementById", "homeMainModal")
	div.Set("innerHTML", c.app.RenderTemplate(
		&songItem, "home/songEdit/index"),
	)

	form := jst.Document.Call("getElementById", "songEditForm")
	form.Call("addEventListener", "submit", c.app.AddEventFuncPreventDefault(c.saveAction))
	cancelButton := jst.Document.Call("getElementById", "songEditCancelButton")
	cancelButton.Call("addEventListener", "click", c.app.AddEventFunc(c.cancelAction))

	albumInput := jst.Document.Call("getElementById", "songEditAlbumInput")
	albumSearchBlock := jst.Document.Call("getElementById", "songEditAlbumSearchBlock")
	albumSearchInput := jst.Document.Call("getElementById", "songEditAlbumSearchInput")
	albumList := jst.Document.Call("getElementById", "songEditAlbumList")

	albumSearchInput.Call("addEventListener", "input", c.app.AddEventFunc(c.albumSearchAction))
	//	albumSearchInput.Call("addEventListener", "focusout", c.app.AddEventFunc(func() {
	//	}))

	albumInput.Call("addEventListener", "focus", c.app.AddEventFunc(func() {
		albumSearchInput.Set("value", "")
		c.albumSearchAction()
		albumInput.Get("style").Set("display", "none")
		albumSearchBlock.Get("style").Set("display", "block")
		albumSearchInput.Call("focus")
	}))

	albumList.Call("addEventListener", "click", c.app.AddRichEventFunc(func(this js.Value, i []js.Value) {
		link := i[0].Get("target").Call("closest", ".albumLink")
		if !link.Truthy() {
			return
		}
		dataset := link.Get("dataset")

		switch link.Get("className").String() {
		case "albumLink":
			albumId := restApiV1.AlbumId(dataset.Get("albumid").String())
			c.songMeta.AlbumId = albumId
			albumInput.Set("value", c.app.localDb.Albums[albumId].Name)

			// Remove searchInput
			albumSearchBlock.Get("style").Set("display", "none")
			albumInput.Get("style").Set("display", "block")
		}
	}))
}

func (c *HomeSongEditComponent) saveAction() {
	if c.closed {
		return
	}

	c.app.ShowLoader("Updating song")

	// Song name
	songName := jst.Document.Call("getElementById", "songEditSongName")
	c.songMeta.Name = songName.Get("value").String()

	// Publication year
	c.songMeta.PublicationYear = nil
	publicationYearStr := jst.Document.Call("getElementById", "songEditPublicationYear").Get("value").String()
	if publicationYearStr != "" {

		publicationYear, err := strconv.ParseInt(publicationYearStr, 10, 64)
		if err == nil {
			c.songMeta.PublicationYear = &publicationYear
		}
	}

	// Album
	// TODO

	// TrackNumber
	c.songMeta.TrackNumber = nil
	trackNumberStr := jst.Document.Call("getElementById", "songEditTrackNumber").Get("value").String()
	if trackNumberStr != "" {

		trackNumber, err := strconv.ParseInt(trackNumberStr, 10, 64)
		if err == nil {
			c.songMeta.TrackNumber = &trackNumber
		}
	}

	// Explicit flag
	c.songMeta.ExplicitFg = jst.Document.Call("getElementById", "songEditExplicitFg").Get("checked").Bool()

	// Artist
	// TODO

	_, cliErr := c.app.restClient.UpdateSong(c.songId, c.songMeta)
	if cliErr != nil {
		c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to update the song", cliErr)
	}

	c.close()
	c.app.HomeComponent.Reload()
	c.app.HideLoader()
}

func (c *HomeSongEditComponent) cancelAction() {
	if c.closed {
		return
	}
	c.close()
}

func (c *HomeSongEditComponent) close() {
	c.closed = true
	c.app.HomeComponent.CloseModal()
}

func (c *HomeSongEditComponent) albumSearchAction() {
	albumSearchInput := jst.Document.Call("getElementById", "songEditAlbumSearchInput")
	nameFilter := albumSearchInput.Get("value").String()

	var resultAlbumList []*restApiV1.Album

	if nameFilter != "" {
		lowerNameFilter := strings.ToLower(nameFilter)
		for _, album := range c.app.localDb.OrderedAlbums {
			if album == nil || !strings.Contains(strings.ToLower(album.Name), lowerNameFilter) {
				continue
			}

			resultAlbumList = append(resultAlbumList, album)
		}
	}

	sort.SliceStable(resultAlbumList, func(i, j int) bool { return len(resultAlbumList[i].Name) < len(resultAlbumList[j].Name) })

	albumList := jst.Document.Call("getElementById", "songEditAlbumList")
	albumList.Set("innerHTML", c.app.RenderTemplate(
		resultAlbumList, "home/songEdit/albumList"),
	)

}
