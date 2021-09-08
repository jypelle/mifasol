package cliwa

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
	message := c.app.doc.Call("getElementById", "message")
	message.Set("innerHTML", msg)
}
