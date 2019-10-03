package cli

import (
	"lyra/cli/ui"
)

func (c *ClientApp) UI() {
	uiApp := ui.NewUIApp(c.config, c.restClient)
	uiApp.Start()
}
