package cliwa

import (
	"bytes"
	"github.com/jypelle/mifasol/internal/cliwa/config"
	"github.com/jypelle/mifasol/internal/cliwa/templates"
	"github.com/jypelle/mifasol/internal/localdb"
	"github.com/jypelle/mifasol/internal/version"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/sirupsen/logrus"
	"html"
	"html/template"
	"strconv"
	"syscall/js"
)

type App struct {
	config     config.ClientConfig
	restClient *restClientV1.RestClient
	localDb    *localdb.LocalDb

	templateHelpers template.FuncMap
	doc             js.Value
}

func NewApp(debugMode bool) *App {

	logrus.Infof("Creation of mifasol web assembly client %s ...", version.AppVersion.String())

	app := &App{
		config: config.ClientConfig{
			ClientEditableConfig: config.NewClientEditableConfig(nil),
		},
		doc: js.Global().Get("document"),
	}

	logrus.Infof("Client created")

	return app
}

func (c *App) Start() {
	c.retrieveServerCredentials()

	c.showStartComponent()

	// Keep goroutine running
	<-make(chan bool)
}

func (c *App) RenderTemplate(content interface{}, filenames ...string) string {

	// Parsing template files
	t := template.New("").Funcs(c.templateHelpers)

	for _, filename := range filenames {

		html, err := templates.Fs.ReadFile(filename)
		if err != nil {
			logrus.Panicf("Unable to read template file %s: %v\n", filename, err)
		}

		t, err = t.Parse(string(html))
		if err != nil {
			logrus.Panicf("Unable to interpret template file %s: %v\n", filename, err)
		}
	}

	var w bytes.Buffer
	err := t.Execute(&w, content)
	if err != nil {
		logrus.Panicf("Unable to execute template files : %v\n", err)
	}

	return w.String()
}

func (c *App) Refresh() {
	if c.localDb == nil {
		return
	}

	err := c.localDb.Refresh()
	if err != nil {
		c.Message("Unable to refresh local database")
	} else {
		c.Message(strconv.Itoa(len(c.localDb.Songs)) + " songs, " + strconv.Itoa(len(c.localDb.Artists)) + " artists, " + strconv.Itoa(len(c.localDb.Albums)) + " albums, " + strconv.Itoa(len(c.localDb.Playlists)) + " playlists ready to be played for " + strconv.Itoa(len(c.localDb.Users)) + " users.")
	}

	artistListDiv := c.doc.Call("getElementById", "artistList")

	var artistListDivContent string
	for _, artist := range c.localDb.OrderedArtists {
		if artist != nil {
			artistListDivContent += "<p>" + html.EscapeString(artist.Name) + "</p>"
		}
	}
	artistListDiv.Set("innerHTML", artistListDivContent)

	albumListDiv := c.doc.Call("getElementById", "albumList")

	var albumListDivContent string
	for _, album := range c.localDb.OrderedAlbums {
		if album != nil {
			albumListDivContent += "<p>" + html.EscapeString(album.Name) + "</p>"
		}
	}
	albumListDiv.Set("innerHTML", albumListDivContent)

	songListDiv := c.doc.Call("getElementById", "songList")

	var songListDivContent string
	for _, song := range c.localDb.OrderedSongs {
		if song != nil {
			songListDivContent += "<p>" + html.EscapeString(song.Name) + "</p>"
		}
	}
	songListDiv.Set("innerHTML", songListDivContent)
}

func (c *App) Message(msg string) {
	message := c.doc.Call("getElementById", "message")
	message.Set("innerHTML", msg)
}
