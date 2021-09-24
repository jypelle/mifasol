package cliwa

import "github.com/jypelle/mifasol/internal/cliwa/jst"

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
