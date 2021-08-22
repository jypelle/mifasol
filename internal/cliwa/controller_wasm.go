package cliwa

import (
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/sirupsen/logrus"
	"net/url"
	"strconv"
	"syscall/js"
)

func (c *App) RetrieveServerCredentials() {
	rawUrl := js.Global().Get("window").Get("location").Get("href").String()
	baseUrl, _ := url.Parse(rawUrl)

	c.config.ServerHostname = baseUrl.Hostname()
	c.config.ServerPort, _ = strconv.ParseInt(baseUrl.Port(), 10, 64)
	c.config.ServerSsl = baseUrl.Scheme == "https"
}

func (c *App) ShowPrehomepage() {
	body := c.doc.Get("body")
	body.Set("innerHTML", `<header style="margin: 2rem;">
		<h1 style="margin:0;">Mifasol</h1>
	</header>
	<main style="display: flex; flex-flow: column nowrap; justify-content: center; align-items: center; flex-grow: 1; overflow-y: auto;">
		<div style="display: flex; flex-flow: row wrap; margin: 2rem; ">
			<div style="margin: 2rem;">
				<h2>Connect to web client (alpha!)</h2>
				<form onsubmit="return logIn()">
						<div>
							<label for="mifasolUsername">Username</label>
							<div>
								<input id="mifasolUsername" type="text" value="">
							</div>
						</div>
						<div>
							<label for="mifasolPassword">Password</label>
							<div>
								<input id="mifasolPassword" type="password" value="">
							</div>
						</div>
						<div>
							<label></label>
							<div>
								<label for="rememberMe">
									<input id="rememberMe" value="true" type="checkbox" checked>
									Remember me
								</label>
							</div>
						</div>
						<p id="message"></p>
						<div>
							<label></label>
							<div>
								<button type="submit">Log in</button>
							</div>
						</div>
				</form>
			</div>
			<div style="margin: 2rem;">
				<h2>Download console client</h2>
				<ul>
					<li><a href="/clients/mifasolcli-windows-amd64.exe">Windows (amd64)</a></li>
					<li><a href="/clients/mifasolcli-linux-amd64">Linux (amd64)</a></li>
					<li><a href="/clients/mifasolcli-linux-arm">Linux (arm)</a></li>
				</ul>
			</div>
		</div>

	</main>`)

	// Set callback
	js.Global().Set("logIn", js.FuncOf(c.logIn))
}

func (c *App) logIn(this js.Value, i []js.Value) interface{} {
	serverUsername := c.doc.Call("getElementById", "mifasolUsername")
	serverPassword := c.doc.Call("getElementById", "mifasolPassword")
	c.config.Username = serverUsername.Get("value").String()
	c.config.Password = serverPassword.Get("value").String()

	go func() {
		// Create rest Client
		var err error
		c.restClient, err = restClientV1.NewRestClient(&c.config, true)
		if err != nil {
			message := c.doc.Call("getElementById", "message")
			message.Set("innerHTML", "Unable to connect to server")
			logrus.Errorf("Unable to instantiate mifasol rest client: %v", err)
			return
		}
		if c.restClient.UserId() == "xxx" {
			message := c.doc.Call("getElementById", "message")
			message.Set("innerHTML", "Wrong credentials")
			return
		}

		c.ShowHomepage()
	}()

	return false
}

func (c *App) ShowHomepage() {
	body := c.doc.Get("body")
	body.Set("innerHTML", `<header style="margin: 2rem;">
		<h1 style="margin:0;">Mifasol</h1>
	</header>
	<main style="display: flex; flex-flow: column nowrap; margin: 2rem; min-height: 0; flex: 1 1 auto; ">
		<p id="message">...</p>
		<div id="artistList" style="display: flex; flex-flow: column nowrap; flex: 1 1 auto; overflow-y: auto;"></div>
		<div>
			<button type="button" onclick="logOut()">Disconnect</button>
		</div>
	</main>
	<footer>
		Playing...
	</footer>`)

	// Set callback
	js.Global().Set("logOut", js.FuncOf(c.logOut))

	go func() {
		artistList, err := c.restClient.ReadArtists(&restApiV1.ArtistFilter{})
		if err != nil {
			logrus.Fatalf("Unable to retrieve artistlist: %v", err)
		}
		artistListDiv := c.doc.Call("getElementById", "artistList")

		for _, artist := range artistList {
			artistElmt := c.doc.Call("createElement", "p")
			artistElmt.Set("innerHTML", artist.Name)
			artistListDiv.Call("appendChild", artistElmt)
		}

		message := c.doc.Call("getElementById", "message")
		message.Set("innerHTML", "Artists loaded")
	}()

}

func (c *App) logOut(this js.Value, i []js.Value) interface{} {
	c.restClient = nil

	c.ShowPrehomepage()

	return nil
}
