package cliwa

import (
	"bytes"
	"github.com/jypelle/mifasol/internal/cliwa/config"
	"github.com/jypelle/mifasol/internal/cliwa/templates"
	"github.com/jypelle/mifasol/internal/localdb"
	"github.com/jypelle/mifasol/internal/version"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/sirupsen/logrus"
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

	messageComponent *MessageComponent
	libraryComponent *LibraryComponent
	currentComponent *CurrentComponent
	playerComponent  *PlayerComponent
}

func NewApp(debugMode bool) *App {

	logrus.Infof("Creation of mifasol web assembly client %s ...", version.AppVersion.String())

	app := &App{
		config: config.ClientConfig{
			ClientEditableConfig: config.NewClientEditableConfig(nil),
		},
		doc: js.Global().Get("document"),
	}

	app.messageComponent = NewMessageComponent(app)
	app.libraryComponent = NewLibraryComponent(app)
	app.currentComponent = NewCurrentComponent(app)
	app.playerComponent = NewPlayerComponent(app)

	logrus.Infof("Client created")

	return app
}

func (c *App) Start() {
	c.retrieveServerCredentials()

	c.showStartComponent()

	// Keep wasm app alive
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

func (c *App) Reload() {
	if c.localDb == nil {
		return
	}
	c.messageComponent.Message("Syncing...")
	// Refresh In memory Db
	err := c.localDb.Refresh()
	if err != nil {
		c.messageComponent.Message("Unable to load data from mifasolsrv")
		return
	}

	c.libraryComponent.RefreshView()

	c.messageComponent.Message(strconv.Itoa(len(c.localDb.Songs)) + " songs, " + strconv.Itoa(len(c.localDb.Artists)) + " artists, " + strconv.Itoa(len(c.localDb.Albums)) + " albums, " + strconv.Itoa(len(c.localDb.Playlists)) + " playlists ready to be played for " + strconv.Itoa(len(c.localDb.Users)) + " users.")
}
