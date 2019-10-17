package cli

import (
	"github.com/jypelle/mifasol/cli/ui"
)

func (c *ClientApp) UI() {
	uiApp := ui.NewUIApp(c.config, c.restClient)
	uiApp.Start()
}
