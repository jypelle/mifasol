package cliwa

import (
	"github.com/jypelle/mifasol/internal/localdb"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/sirupsen/logrus"
	"syscall/js"
)

func (c *App) logInAction(this js.Value, i []js.Value) interface{} {
	serverUsername := c.doc.Call("getElementById", "mifasolUsername")
	serverPassword := c.doc.Call("getElementById", "mifasolPassword")
	c.config.Username = serverUsername.Get("value").String()
	c.config.Password = serverPassword.Get("value").String()

	go func() {
		// Create rest Client
		restClient, err := restClientV1.NewRestClient(&c.config, true)
		if err != nil {
			message := c.doc.Call("getElementById", "message")
			message.Set("innerHTML", "Unable to connect to server")
			logrus.Errorf("Unable to instantiate mifasol rest client: %v", err)
			return
		}
		if restClient.UserId() == "xxx" {
			message := c.doc.Call("getElementById", "message")
			message.Set("innerHTML", "Wrong credentials")
			return
		}

		c.restClient = restClient
		c.localDb = localdb.NewLocalDb(c.restClient, c.config.Collator())

		c.showHomeComponent()
	}()

	return false
}

func (c *App) logOutAction(this js.Value, i []js.Value) interface{} {
	c.showStartComponent()
	return nil
}

func (c *App) refreshAction(this js.Value, i []js.Value) interface{} {
	go c.Reload()
	return nil
}

func (c *App) playSong(songId restApiV1.SongId) {
	token, cliErr := c.restClient.GetToken()

	if cliErr != nil {
		return
	}

	player := c.doc.Call("getElementById", "player")
	player.Set("src", "/api/v1/songContents/"+string(songId)+"?bearer="+token.AccessToken)
	player.Call("play")

	return
}
