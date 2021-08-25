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
	js.Global().Set("logInAction", js.FuncOf(c.logInAction))
}

func (c *App) showHomeComponent() {
	body := c.doc.Get("body")
	body.Set("innerHTML", c.RenderTemplate(nil, "home.html"))

	// Set callback
	js.Global().Set("logOutAction", js.FuncOf(c.logOutAction))
	js.Global().Set("refreshAction", js.FuncOf(c.refreshAction))

	go func() {
		c.Refresh()
	}()

}
