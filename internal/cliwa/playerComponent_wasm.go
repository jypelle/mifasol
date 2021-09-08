package cliwa

type PlayerComponent struct {
	app *App
}

func NewPlayerComponent(app *App) *PlayerComponent {
	c := &PlayerComponent{
		app: app,
	}

	return c
}
