package cliwa

import (
	"github.com/jypelle/mifasol/internal/localdb"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/sirupsen/logrus"
)

type StartComponent struct {
	app *App
}

func NewStartComponent(app *App) *StartComponent {
	c := &StartComponent{
		app: app,
	}

	return c
}

func (c *StartComponent) Show() {
	c.app.restClient = nil
	c.app.localDb = nil

	body := c.app.doc.Get("body")
	body.Set("innerHTML", c.app.RenderTemplate(nil, "start.html"))

	// Set focus
	c.app.doc.Call("getElementById", "mifasolUsername").Call("focus")

	// Set button
	//js.Global().Set("logInAction", c.app.AddEventFunc(c.logInAction))
	startForm := c.app.doc.Call("getElementById", "startForm")
	startForm.Call("addEventListener", "submit", c.app.AddEventFuncPreventDefault(c.logInAction))
}

func (c *StartComponent) logInAction() {
	serverUsername := c.app.doc.Call("getElementById", "mifasolUsername")
	serverPassword := c.app.doc.Call("getElementById", "mifasolPassword")
	c.app.config.Username = serverUsername.Get("value").String()
	c.app.config.Password = serverPassword.Get("value").String()

	// Create rest Client
	restClient, err := restClientV1.NewRestClient(&c.app.config, true)
	if err != nil {
		message := c.app.doc.Call("getElementById", "message")
		message.Set("innerHTML", "Unable to connect to server")
		logrus.Errorf("Unable to instantiate mifasol rest client: %v", err)
		return
	}
	if restClient.UserId() == "xxx" {
		message := c.app.doc.Call("getElementById", "message")
		message.Set("innerHTML", "Wrong credentials")
		return
	}

	c.app.restClient = restClient
	c.app.localDb = localdb.NewLocalDb(c.app.restClient, c.app.config.Collator())

	c.app.HomeComponent = NewHomeComponent(c.app)
	c.app.HomeComponent.Show()
}
