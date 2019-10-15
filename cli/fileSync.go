package cli

import (
	"mifasol/cli/fileSync"
)

func (c *ClientApp) FileSyncInit(fileSyncMusicFolder string) {

	fileSyncApp := fileSync.NewFileSyncApp(c.config, c.restClient, fileSyncMusicFolder)
	fileSyncApp.Init()

}

func (c *ClientApp) FileSyncSync(fileSyncMusicFolder string) {

	fileSyncApp := fileSync.NewFileSyncApp(c.config, c.restClient, fileSyncMusicFolder)
	fileSyncApp.Sync()

}
