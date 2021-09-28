package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restClientV1"
)

type MessageComponent struct {
	app *App
}

func NewMessageComponent(app *App) *MessageComponent {
	c := &MessageComponent{
		app: app,
	}

	return c
}

func (c *MessageComponent) Message(msg string) {
	message := jst.Document.Call("getElementById", "message")
	message.Set("innerHTML", msg)
}

func (c *MessageComponent) WarningMessage(msg string) {
	c.Message(`<span style="color: red;">` + msg + `</span>`)
}

func (c *MessageComponent) ClientErrorMessage(message string, cliErr restClientV1.ClientError) {
	c.WarningMessage(message + " (" + cliErr.Code().String() + ")")
}
