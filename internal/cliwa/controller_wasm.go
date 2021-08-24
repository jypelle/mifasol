package cliwa

import (
	"github.com/jypelle/mifasol/internal/localdb"
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
	body.Set("innerHTML", `<main style="display: flex; flex-flow: column nowrap; justify-content: center; align-items: center; flex-grow: 1;">
		<h1 style="margin:0; margin: 2rem;"><img src="/static/image/logo64.png" style="vertical-align:middle;"> Mifasol</h1>
		<div style="display: flex; flex-flow: row wrap;">
			<div style="margin: 2rem;">
				<h2>Connect to web client (alpha!)</h2>
				<form onsubmit="return logIn()">
						<div>
							<label for="mifasolUsername">Username</label>
							<div>
								<input id="mifasolUsername" type="text">
							</div>
						</div>
						<div>
							<label for="mifasolPassword">Password</label>
							<div>
								<input id="mifasolPassword" type="password">
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

	// Set focus
	c.doc.Call("getElementById", "mifasolUsername").Call("focus")

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

		c.localDb = localdb.NewLocalDb(c.restClient, c.config.Collator())

		c.ShowHomepage()
	}()

	return false
}

func (c *App) ShowHomepage() {
	body := c.doc.Get("body")
	body.Set("innerHTML", `<header style="display: flex; flex-flow: row wrap; justify-content: space-between;">
		<h1 style="margin: 1rem;"><img src="/static/image/logo32.png" style="vertical-align:middle;"> Mifasol</h1>
		<button style="margin: 1rem;" type="button" onclick="logOut()">Log out</button>
	</header>
	<main style="display: flex; flex-flow: column nowrap; margin: 0 2rem 0 2rem; flex: 1 1 auto; min-height: 0;">
		<div style="display: flex; flex-flow: column wrap; flex: 1 1 auto; min-height: 0; margin: -2rem 0 0 -2rem;">
			<div style="display: flex; flex-flow: column nowrap; flex: 1 1 auto; min-height: 0; margin: 2rem 0 0 2rem;">
				<h2 style="margin:0 0 1rem 0;">Artists</h2>
				<div id="artistList" style="display: flex; flex-flow: column nowrap; flex: 1 1 auto; overflow-y: auto;"></div>
			</div>
			<div style="display: flex; flex-flow: column nowrap; flex: 1 1 auto; min-height: 0; margin: 2rem 0 0 2rem;">
				<h2 style="margin:0 0 1rem 0;">Albums</h2>
				<div id="albumList" style="display: flex; flex-flow: column nowrap; flex: 1 1 auto; overflow-y: auto;"></div>
			</div>
			<div style="display: flex; flex-flow: column nowrap; flex: 1 1 auto; min-height: 0; margin: 2rem 0 0 2rem;">
				<h2 style="margin:0 0 1rem 0;">Songs</h2>
				<div id="songList" style="display: flex; flex-flow: column nowrap; flex: 1 1 auto; overflow-y: auto;"></div>
			</div>
		</div>
	</main>
	<footer style="display: flex; flex-flow: row wrap;">
		<p style="margin: 1rem;" id="message">...</p>
	</footer>`)

	// Set callback
	js.Global().Set("logOut", js.FuncOf(c.logOut))

	go func() {
		err := c.localDb.Refresh()
		if err != nil {
			logrus.Fatalf("Unable to refresh local db: %v", err)
		}

		artistListDiv := c.doc.Call("getElementById", "artistList")

		for _, artist := range c.localDb.OrderedArtists {
			if artist != nil {
				artistElmt := c.doc.Call("createElement", "p")
				artistElmt.Set("innerHTML", artist.Name)
				artistListDiv.Call("appendChild", artistElmt)
			}
		}

		albumListDiv := c.doc.Call("getElementById", "albumList")

		for _, album := range c.localDb.OrderedAlbums {
			if album != nil {
				albumElmt := c.doc.Call("createElement", "p")
				albumElmt.Set("innerHTML", album.Name)
				albumListDiv.Call("appendChild", albumElmt)
			}
		}

		songListDiv := c.doc.Call("getElementById", "songList")

		for _, song := range c.localDb.OrderedSongs {
			if song != nil {
				songElmt := c.doc.Call("createElement", "p")
				songElmt.Set("innerHTML", song.Name)
				songListDiv.Call("appendChild", songElmt)
			}
		}

		//message := c.doc.Call("getElementById", "message")
		//message.Set("innerHTML", "Artists loaded")
	}()

}

func (c *App) logOut(this js.Value, i []js.Value) interface{} {
	c.restClient = nil

	c.ShowPrehomepage()

	return nil
}
