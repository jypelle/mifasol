package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restClientV1"
)

type HomeMessageComponent struct {
	app *App
}

func NewHomeMessageComponent(app *App) *HomeMessageComponent {
	c := &HomeMessageComponent{
		app: app,
	}

	return c
}

func (c *HomeMessageComponent) Message(msg string) {
	message := jst.Id("message")
	message.Set("innerHTML", msg)
}

func (c *HomeMessageComponent) WarningMessage(msg string) {
	c.Message(`<span style="color: red;">` + msg + `</span>`)
}

func (c *HomeMessageComponent) ClientErrorMessage(message string, cliErr restClientV1.ClientError) {
	c.WarningMessage(message + " (" + cliErr.Code().String() + ")")
}
