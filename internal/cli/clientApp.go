package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/jypelle/mifasol/internal/cli/config"
	"github.com/jypelle/mifasol/internal/cli/fileSync"
	"github.com/jypelle/mifasol/internal/cli/imp"
	"github.com/jypelle/mifasol/internal/cli/ui"
	"github.com/jypelle/mifasol/internal/version"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
)

type ClientApp struct {
	config     config.ClientConfig
	restClient *restClientV1.RestClient
}

func NewClientApp(configDir string, debugMode bool) *ClientApp {

	logrus.Debugf("Creation of mifasol client %s ...", version.AppVersion.String())

	app := &ClientApp{
		config: config.ClientConfig{
			ConfigDir: configDir,
			DebugMode: debugMode,
		},
	}

	// Check Configuration folder
	_, err := os.Stat(app.config.ConfigDir)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Printf("Creation of config folder: %s", app.config.ConfigDir)
			err = os.Mkdir(app.config.ConfigDir, 0770)
			if err != nil {
				logrus.Fatalf("Unable to create config folder: %v\n", err)
			}
		} else {
			logrus.Fatalf("Unable to access config folder: %s", app.config.ConfigDir)
		}
	}

	// Open configuration file
	var draftClientEditableConfig *config.ClientEditableConfig

	rawConfig, err := ioutil.ReadFile(app.config.GetCompleteConfigFilename())
	if err == nil {
		// Interpret configuration file
		draftClientEditableConfig = &config.ClientEditableConfig{}
		err = json.Unmarshal(rawConfig, draftClientEditableConfig)
		if err != nil {
			logrus.Fatalf("Unable to interpret config file: %v\n", err)
		}
	}

	app.config.ClientEditableConfig = config.NewClientEditableConfig(draftClientEditableConfig)

	app.config.Save()

	logrus.Debugln("Client created")

	return app
}

func (c *ClientApp) Init() {

	// Create rest Client
	var e error
	c.restClient, e = restClientV1.NewRestClient(&c.config)
	if e != nil {
		if e == restClientV1.ErrBadCertificate {
			fmt.Print("Mifasol server certificate has changed.\nDo you accept the new one (to prevent man-in-the-middle attack, you should explicitely accept) ?[y/N]\n")
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			text = strings.Replace(text, "\n", "", -1)
			text = strings.Replace(text, "\r", "", -1)
			if text == "y" || text == "Y" {
				os.Remove(c.config.GetCompleteConfigCertFilename())
				c.Init()
				return
			} else {
				fmt.Print("New server certificate has been refused.")
				os.Exit(1)
			}
		}
		logrus.Fatalf("Unable to instanciate mifasol rest client: %v\n", e)
	}
}

func (c *ClientApp) FileSyncInit(fileSyncMusicFolder string) {

	fileSyncApp := fileSync.NewApp(c.config, c.restClient, fileSyncMusicFolder)
	fileSyncApp.Init()

}

func (c *ClientApp) FileSyncSync(fileSyncMusicFolder string) {

	fileSyncApp := fileSync.NewApp(c.config, c.restClient, fileSyncMusicFolder)
	fileSyncApp.Sync()

}

func (c *ClientApp) Import(importDir string, importOneFolderPerAlbumDisabled bool) {

	importApp := imp.NewApp(c.config, c.restClient, importDir, importOneFolderPerAlbumDisabled)
	importApp.Start()
}

func (c *ClientApp) UI() {
	uiApp := ui.NewApp(c.config, c.restClient)
	uiApp.Start()
}
