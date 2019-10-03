package cli

import "lyra/cli/imp"

func (c *ClientApp) Import(importDir string) {

	importApp := imp.NewImpApp(c.config, c.restClient, importDir)
	importApp.Start()
}
