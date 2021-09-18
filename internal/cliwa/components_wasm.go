package cliwa

import (
	"net/url"
	"strconv"
	"syscall/js"
)

func (c *App) retrieveServerCredentials() {
	rawUrl := js.Global().Get("window").Get("location").Get("href").String()
	baseUrl, _ := url.Parse(rawUrl)

	c.config.ServerHostname = baseUrl.Hostname()
	c.config.ServerPort, _ = strconv.ParseInt(baseUrl.Port(), 10, 64)
	c.config.ServerSsl = baseUrl.Scheme == "https"
}

func (c *App) showStartComponent() {
	c.restClient = nil
	c.localDb = nil

	body := c.doc.Get("body")
	body.Set("innerHTML", c.RenderTemplate(nil, "start.html"))

	// Set focus
	c.doc.Call("getElementById", "mifasolUsername").Call("focus")

	// Set button
	js.Global().Set("logInAction", c.AddEventFunc(c.logInAction))
}

func (c *App) showHomeComponent() {
	body := c.doc.Get("body")
	body.Set("innerHTML", c.RenderTemplate(nil, "home.html"))

	// Set buttons
	logOutButton := c.doc.Call("getElementById", "logOutButton")
	logOutButton.Call("addEventListener", "click", c.AddEventFunc(c.showStartComponent))
	refreshButton := c.doc.Call("getElementById", "refreshButton")
	refreshButton.Call("addEventListener", "click", c.AddEventFunc(c.refreshAction))

	c.libraryComponent.Show()
	c.currentComponent.Show()

	c.Reload()

}
