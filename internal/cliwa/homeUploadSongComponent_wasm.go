package cliwa

import (
	"bytes"
	"fmt"
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restApiV1"
	"html"
	"strings"
	"syscall/js"
)

type HomeUploadSongComponent struct {
	app          *App
	closed       bool
	songFiles    []js.Value
	songFilesIdx int
}

func NewHomeUploadSongComponent(app *App) *HomeUploadSongComponent {
	c := &HomeUploadSongComponent{
		app: app,
	}

	return c
}

func (c *HomeUploadSongComponent) Show() {
	div := jst.Id("homeMainModal")
	div.Set("innerHTML", c.app.RenderTemplate(
		nil, "home/uploadSong/index"),
	)

	form := jst.Id("uploadSongForm")
	form.Call("addEventListener", "submit", c.app.AddEventFuncPreventDefault(c.uploadAction))
	cancelButton := jst.Id("uploadSongCancelButton")
	cancelButton.Call("addEventListener", "click", c.app.AddEventFunc(c.cancelAction))
	uploadSongFolder := jst.Id("uploadSongFolder")
	uploadSongFolder.Call("addEventListener", "change", c.app.AddEventFunc(func() {
		files := uploadSongFolder.Get("files")

		// Keep only flac and mp3 files
		c.songFiles = nil
		c.songFilesIdx = 0
		for i := 0; i < files.Length(); i++ {
			file := files.Index(i)
			lowerName := strings.ToLower(file.Get("name").String())
			if strings.HasSuffix(lowerName, ".mp3") || strings.HasSuffix(lowerName, ".flac") {
				c.songFiles = append(c.songFiles, file)
			}
		}

		jst.Id("uploadSongReport").Set("innerHTML", fmt.Sprintf("%d song(s) to import", len(c.songFiles)))

	}))
}

func (c *HomeUploadSongComponent) uploadAction() {
	if c.closed {
		return
	}

	if len(c.songFiles) == 0 {
		return
	}

	div := jst.Id("homeMainModal")
	div.Set("innerHTML", c.app.RenderTemplate(
		nil, "home/uploadSong/uploading"),
	)
	uploadingCancelButton := jst.Id("uploadSongUploadingCancelButton")
	uploadingCancelButton.Call("addEventListener", "click", c.app.AddEventFunc(c.uploadingCancelAction))

	c.loadSongAction()
}

func (c *HomeUploadSongComponent) loadSongAction() {
	if c.closed {
		return
	}
	songFile := c.songFiles[c.songFilesIdx]
	uploadingStatus := jst.Id("uploadSongUploadingStatus")
	uploadingStatus.Set("innerHTML", fmt.Sprintf("%s", html.EscapeString(songFile.Get("webkitRelativePath").String())))
	reader := js.Global().Get("FileReader").New()
	reader.Call("addEventListener", "load", c.app.AddRichEventFunc(c.uploadSongAction))
	reader.Call("readAsArrayBuffer", songFile)
}

func (c *HomeUploadSongComponent) uploadSongAction(this js.Value, args []js.Value) {
	if c.closed {
		return
	}
	songFile := c.songFiles[c.songFilesIdx]
	uploadingStatus := jst.Id("uploadSongUploadingStatus")
	uploadingStatus.Set("innerHTML", fmt.Sprintf("%s", html.EscapeString(songFile.Get("webkitRelativePath").String())))

	result := this.Get("result")
	jscontent := js.Global().Get("Uint8Array").New(result)
	content := make([]byte, songFile.Get("size").Int())
	js.CopyBytesToGo(content, jscontent)

	lowerName := strings.ToLower(songFile.Get("name").String())
	var songFormat restApiV1.SongFormat
	if strings.HasSuffix(lowerName, ".flac") {
		songFormat = restApiV1.SongFormatFlac
	} else {
		songFormat = restApiV1.SongFormatMp3
	}

	_, cliErr := c.app.restClient.CreateSongContent(songFormat, bytes.NewReader(content))
	if cliErr != nil {
		c.app.HomeComponent.MessageComponent.ClientErrorMessage(fmt.Sprintf("Unable to upload song %s", songFile.Get("name")), cliErr)
	}

	uploadSongUploadingProgressbar := jst.Id("uploadSongUploadingProgressbar")
	uploadSongUploadingProgressbar.Get("style").Set("width", fmt.Sprintf("%f%%", 100.0*float64(c.songFilesIdx+1)/float64(len(c.songFiles))))

	if c.songFilesIdx < len(c.songFiles)-1 {
		c.songFilesIdx++
		c.loadSongAction()
	} else {
		c.close()
		c.app.HomeComponent.Reload()
	}
}

func (c *HomeUploadSongComponent) cancelAction() {
	if c.closed {
		return
	}
	c.close()
}

func (c *HomeUploadSongComponent) uploadingCancelAction() {
	if c.closed {
		return
	}
	c.close()
	c.app.HomeComponent.Reload()
}

func (c *HomeUploadSongComponent) close() {
	c.closed = true
	c.app.HomeComponent.CloseModal()
}
