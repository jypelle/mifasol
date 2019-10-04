package cli

import "lyra/cli/imp"

func (c *ClientApp) Import(importDir string, importOneFolderPerAlbumDisabled bool) {

	importApp := imp.NewImpApp(c.config, c.restClient, importDir, importOneFolderPerAlbumDisabled)
	importApp.Start()
}
