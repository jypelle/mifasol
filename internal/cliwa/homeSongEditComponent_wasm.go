package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restApiV1"
	"html"
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

	div := jst.Id("homeMainModal")
	div.Set("innerHTML", c.app.RenderTemplate(
		&songItem, "home/songEdit/index"),
	)

	form := jst.Id("songEditForm")
	form.Call("addEventListener", "submit", c.app.AddEventFuncPreventDefault(c.saveAction))
	cancelButton := jst.Id("songEditCancelButton")
	cancelButton.Call("addEventListener", "click", c.app.AddEventFunc(c.cancelAction))

	// Album

	albumCurrentBlock := jst.Id("songEditAlbumCurrentBlock")
	albumCurrentName := jst.Id("songEditAlbumCurrentName")
	albumCurrentDelete := jst.Id("songEditAlbumCurrentDelete")
	albumSearchBlock := jst.Id("songEditAlbumSearchBlock")
	albumSearchInput := jst.Id("songEditAlbumSearchInput")
	albumSearchClean := jst.Id("songEditAlbumSearchClean")
	albumSearchList := jst.Id("songEditAlbumSearchList")

	albumCurrentDelete.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		c.songMeta.AlbumId = restApiV1.UnknownAlbumId
		albumCurrentName.Set("innerHTML", "")
		// Add searchInput
		albumSearchBlock.Get("style").Set("display", "block")
		albumCurrentBlock.Get("style").Set("display", "none")
	}))

	albumSearchInput.Call("addEventListener", "input", c.app.AddEventFunc(c.albumSearchAction))
	albumSearchClean.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		// Clear
		albumSearchInput.Set("value", "")
		c.albumSearchAction()
	}))
	albumSearchList.Call("addEventListener", "click", c.app.AddRichEventFunc(func(this js.Value, i []js.Value) {
		link := i[0].Get("target").Call("closest", ".albumLink")
		if !link.Truthy() {
			return
		}
		dataset := link.Get("dataset")

		switch link.Get("className").String() {
		case "albumLink":
			albumId := restApiV1.AlbumId(dataset.Get("albumid").String())
			c.songMeta.AlbumId = albumId
			albumCurrentName.Set("innerHTML", html.EscapeString(c.app.localDb.Albums[albumId].Name))

			// Clear
			albumSearchInput.Set("value", "")
			c.albumSearchAction()

			// Remove searchInput
			albumSearchBlock.Get("style").Set("display", "none")
			albumCurrentBlock.Get("style").Set("display", "flex")
		}
	}))

	// Artists
	artistCurrentList := jst.Id("songEditArtistCurrentList")
	artistSearchInput := jst.Id("songEditArtistSearchInput")
	artistSearchClean := jst.Id("songEditArtistSearchClean")
	artistSearchList := jst.Id("songEditArtistSearchList")

	// Remove artist
	artistCurrentList.Call("addEventListener", "click", c.app.AddRichEventFunc(func(this js.Value, i []js.Value) {
		link := i[0].Get("target").Call("closest", ".artistLink")
		if !link.Truthy() {
			return
		}
		dataset := link.Get("dataset")

		switch link.Get("className").String() {
		case "artistLink":
			artistId := restApiV1.ArtistId(dataset.Get("artistid").String())

			for idx, songArtistId := range c.songMeta.ArtistIds {
				if songArtistId == artistId {
					if idx == len(c.songMeta.ArtistIds)-1 {
						c.songMeta.ArtistIds = c.songMeta.ArtistIds[0:idx]
					} else {
						c.songMeta.ArtistIds = append(c.songMeta.ArtistIds[0:idx], c.songMeta.ArtistIds[idx+1])
					}

					break
				}
			}

			// Refresh current artists
			c.refreshCurrentArtistAction()
		}
	}))

	// Search artist
	artistSearchInput.Call("addEventListener", "input", c.app.AddEventFunc(c.artistSearchAction))
	artistSearchClean.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		// Clear search input
		artistSearchInput.Set("value", "")
		c.artistSearchAction()
	}))

	// Add artist
	artistSearchList.Call("addEventListener", "click", c.app.AddRichEventFunc(func(this js.Value, i []js.Value) {
		link := i[0].Get("target").Call("closest", ".artistLink")
		if !link.Truthy() {
			return
		}
		dataset := link.Get("dataset")

		switch link.Get("className").String() {
		case "artistLink":
			artistId := restApiV1.ArtistId(dataset.Get("artistid").String())
			c.songMeta.ArtistIds = append(c.songMeta.ArtistIds, artistId)

			// Clear search input
			artistSearchInput.Set("value", "")
			c.artistSearchAction()

			// Refresh current artists
			c.refreshCurrentArtistAction()
		}
	}))

	c.refreshCurrentArtistAction()

}

func (c *HomeSongEditComponent) saveAction() {
	if c.closed {
		return
	}

	c.app.ShowLoader("Updating song")

	// Song name
	songName := jst.Id("songEditSongName")
	c.songMeta.Name = songName.Get("value").String()

	// Publication year
	c.songMeta.PublicationYear = nil
	publicationYearStr := jst.Id("songEditPublicationYear").Get("value").String()
	if publicationYearStr != "" {

		publicationYear, err := strconv.ParseInt(publicationYearStr, 10, 64)
		if err == nil {
			c.songMeta.PublicationYear = &publicationYear
		}
	}

	// TrackNumber
	c.songMeta.TrackNumber = nil
	trackNumberStr := jst.Id("songEditTrackNumber").Get("value").String()
	if trackNumberStr != "" {

		trackNumber, err := strconv.ParseInt(trackNumberStr, 10, 64)
		if err == nil {
			c.songMeta.TrackNumber = &trackNumber
		}
	}

	// Explicit flag
	c.songMeta.ExplicitFg = jst.Id("songEditExplicitFg").Get("checked").Bool()

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
	albumSearchInput := jst.Id("songEditAlbumSearchInput")
	albumSearchList := jst.Id("songEditAlbumSearchList")
	nameFilter := albumSearchInput.Get("value").String()

	type AlbumSearchItem struct {
		AlbumId        restApiV1.AlbumId
		AlbumName      string
		AlbumSongCount int
		Artists        []struct {
			ArtistId   string
			ArtistName string
		}
	}

	var resultAlbumList []*AlbumSearchItem

	if nameFilter != "" {
		lowerNameFilter := strings.ToLower(nameFilter)
		for _, album := range c.app.localDb.OrderedAlbums {
			if album == nil || !strings.Contains(strings.ToLower(album.Name), lowerNameFilter) {
				continue
			}

			albumSearchItem := &AlbumSearchItem{
				AlbumId:        album.Id,
				AlbumName:      album.Name,
				AlbumSongCount: len(c.app.localDb.AlbumOrderedSongs[album.Id]),
			}
			for _, artistId := range album.ArtistIds {
				albumSearchItem.Artists = append(albumSearchItem.Artists, struct {
					ArtistId   string
					ArtistName string
				}{
					ArtistId:   string(artistId),
					ArtistName: c.app.localDb.Artists[artistId].Name,
				})
			}

			resultAlbumList = append(resultAlbumList, albumSearchItem)
		}

		sort.SliceStable(resultAlbumList, func(i, j int) bool {
			return len(resultAlbumList[i].AlbumName) < len(resultAlbumList[j].AlbumName)
		})

		if len(resultAlbumList) > 100 {
			resultAlbumList = resultAlbumList[0:100]
		}

		albumSearchList.Set("innerHTML", c.app.RenderTemplate(
			resultAlbumList, "home/songEdit/albumSearchList"),
		)
		albumSearchList.Get("style").Set("display", "block")
	} else {
		albumSearchList.Set("innerHTML", "")
		albumSearchList.Get("style").Set("display", "none")
	}

}

func (c *HomeSongEditComponent) refreshCurrentArtistAction() {
	type ArtistCurrentItem struct {
		ArtistId   restApiV1.ArtistId
		ArtistName string
	}

	var resultArtistList []*ArtistCurrentItem

	for _, artistId := range c.songMeta.ArtistIds {
		artistCurrentItem := &ArtistCurrentItem{
			ArtistId:   artistId,
			ArtistName: c.app.localDb.Artists[artistId].Name,
		}

		resultArtistList = append(resultArtistList, artistCurrentItem)
	}

	artistCurrentList := jst.Id("songEditArtistCurrentList")
	artistCurrentList.Set("innerHTML", c.app.RenderTemplate(
		resultArtistList, "home/songEdit/artistCurrentList"),
	)
}

func (c *HomeSongEditComponent) artistSearchAction() {
	artistSearchInput := jst.Id("songEditArtistSearchInput")
	artistSearchList := jst.Id("songEditArtistSearchList")

	nameFilter := artistSearchInput.Get("value").String()

	type ArtistSearchItem struct {
		ArtistId        restApiV1.ArtistId
		ArtistName      string
		ArtistSongCount int
	}

	var resultArtistList []*ArtistSearchItem

	if nameFilter != "" {
		lowerNameFilter := strings.ToLower(nameFilter)
		for _, artist := range c.app.localDb.OrderedArtists {

			if artist == nil || !strings.Contains(strings.ToLower(artist.Name), lowerNameFilter) {
				continue
			}

			artistOfCurrentSong := false
			for _, songArtistId := range c.songMeta.ArtistIds {
				if artist.Id == songArtistId {
					artistOfCurrentSong = true
					break
				}
			}
			if artistOfCurrentSong {
				continue
			}

			artistSearchItem := &ArtistSearchItem{
				ArtistId:        artist.Id,
				ArtistName:      artist.Name,
				ArtistSongCount: len(c.app.localDb.ArtistOrderedSongs[artist.Id]),
			}

			resultArtistList = append(resultArtistList, artistSearchItem)
		}

		sort.SliceStable(resultArtistList, func(i, j int) bool {
			return len(resultArtistList[i].ArtistName) < len(resultArtistList[j].ArtistName)
		})

		if len(resultArtistList) > 100 {
			resultArtistList = resultArtistList[0:100]
		}

		artistSearchList.Set("innerHTML", c.app.RenderTemplate(
			resultArtistList, "home/songEdit/artistSearchList"),
		)
		artistSearchList.Get("style").Set("display", "block")
	} else {
		artistSearchList.Set("innerHTML", "")
		artistSearchList.Get("style").Set("display", "none")
	}
}
