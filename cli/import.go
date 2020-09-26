package cli

import "github.com/jypelle/mifasol/cli/imp"

func (c *ClientApp) Import(importDir string, importOneFolderPerAlbumDisabled bool) {

	importApp := imp.NewApp(c.config, c.restClient, importDir, importOneFolderPerAlbumDisabled)
	importApp.Start()
}
