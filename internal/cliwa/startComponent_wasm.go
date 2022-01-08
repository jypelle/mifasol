package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
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

func (c *StartComponent) Render() {
	// No autolog or autolog failed
	mainComponent := jst.Id("mainComponent")
	mainComponent.Set("innerHTML", c.app.RenderTemplate(nil, "start/index"))

	// Set focus
	jst.Id("mifasolUsername").Call("focus")

	// Set button
	startForm := jst.Id("startForm")
	startForm.Call("addEventListener", "submit", c.app.AddEventFuncPreventDefault(c.logInAction))
}

func (c *StartComponent) logInAction() {
	serverUsername := jst.Id("mifasolUsername")
	serverPassword := jst.Id("mifasolPassword")
	c.app.config.Username = serverUsername.Get("value").String()
	c.app.config.Password = serverPassword.Get("value").String()

	// Create rest Client
	var err error
	c.app.restClient, err = restClientV1.NewRestClient(&c.app.config, true)
	messageBlock := jst.Id("startMessageBlock")
	message := jst.Id("startMessage")
	if err != nil {
		messageBlock.Get("style").Set("display", "flex")
		message.Set("innerHTML", "Unable to connect to server")
		logrus.Errorf("Unable to instantiate mifasol rest client: %v", err)
		return
	}
	if c.app.ConnectedUserId() == restApiV1.UndefinedUserId {
		messageBlock.Get("style").Set("display", "flex")
		message.Set("innerHTML", "Wrong credentials")
		jst.LocalStorage.Set("mifasolUsername", "")
		jst.LocalStorage.Set("mifasolPassword", "")
		return
	}

	rememberMe := jst.Id("rememberMe").Get("checked").Bool()

	if rememberMe {
		// Store user & password in localStorage
		jst.LocalStorage.Set("mifasolUsername", c.app.config.Username)
		jst.LocalStorage.Set("mifasolPassword", c.app.config.Password)
	}

	c.app.ConnectAction()
}
