package cli

import (
	"mifasol/cli/ui"
)

func (c *ClientApp) UI() {
	uiApp := ui.NewUIApp(c.config, c.restClient)
	uiApp.Start()
}
