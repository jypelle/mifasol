package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/config"
	"github.com/jypelle/mifasol/internal/localdb"
	"github.com/jypelle/mifasol/internal/version"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/sirupsen/logrus"
	"syscall/js"
)

type App struct {
	config     config.ClientConfig
	restClient *restClientV1.RestClient
	localDb    *localdb.LocalDb
	doc        js.Value
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
	c.RetrieveServerCredentials()

	c.ShowPrehomepage()

	// Keep goroutine running
	<-make(chan bool)
}
