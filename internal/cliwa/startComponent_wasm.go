package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/internal/localdb"
	"github.com/jypelle/mifasol/restApiV1"
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

	// Autolog ?
	username := jst.LocalStorage.Get("mifasolUsername").String()
	password := jst.LocalStorage.Get("mifasolPassword").String()
	if username != "" || password != "" {
		c.app.config.Username = username
		c.app.config.Password = password

		// Create rest Client
		var err error
		c.app.restClient, err = restClientV1.NewRestClient(&c.app.config, true)
		if err != nil {
			logrus.Errorf("Unable to instantiate mifasol rest client: %v", err)
		} else {
			if c.app.ConnectedUserId() == restApiV1.UndefinedUserId {
				logrus.Errorf("Wrong credentials")

				jst.LocalStorage.Set("mifasolUsername", "")
				jst.LocalStorage.Set("mifasolPassword", "")
			} else {
				c.goHome()
				return
			}
		}
	}

	// No autolog or autolog failed
	mainComponent := jst.Document.Call("getElementById", "mainComponent")
	mainComponent.Set("innerHTML", c.app.RenderTemplate(nil, "start.html"))

	// Set focus
	jst.Document.Call("getElementById", "mifasolUsername").Call("focus")

	// Set button
	//js.Global().Set("logInAction", c.app.AddEventFunc(c.logInAction))
	startForm := jst.Document.Call("getElementById", "startForm")
	startForm.Call("addEventListener", "submit", c.app.AddEventFuncPreventDefault(c.logInAction))

}

func (c *StartComponent) logInAction() {
	serverUsername := jst.Document.Call("getElementById", "mifasolUsername")
	serverPassword := jst.Document.Call("getElementById", "mifasolPassword")
	c.app.config.Username = serverUsername.Get("value").String()
	c.app.config.Password = serverPassword.Get("value").String()

	// Create rest Client
	var err error
	c.app.restClient, err = restClientV1.NewRestClient(&c.app.config, true)
	if err != nil {
		message := jst.Document.Call("getElementById", "message")
		message.Set("innerHTML", "Unable to connect to server")
		logrus.Errorf("Unable to instantiate mifasol rest client: %v", err)
		return
	}
	if c.app.ConnectedUserId() == restApiV1.UndefinedUserId {
		message := jst.Document.Call("getElementById", "message")
		message.Set("innerHTML", "Wrong credentials")
		jst.LocalStorage.Set("mifasolUsername", "")
		jst.LocalStorage.Set("mifasolPassword", "")
		return
	}

	rememberMe := jst.Document.Call("getElementById", "rememberMe").Get("checked").Bool()

	if rememberMe {
		// Store user & password in localStorage
		jst.LocalStorage.Set("mifasolUsername", c.app.config.Username)
		jst.LocalStorage.Set("mifasolPassword", c.app.config.Password)
	}

	c.goHome()
}

func (c *StartComponent) goHome() {
	//	c.app.restClient = restClient
	c.app.localDb = localdb.NewLocalDb(c.app.restClient, c.app.config.Collator())

	c.app.HomeComponent = NewHomeComponent(c.app)
	c.app.HomeComponent.Show()

}
