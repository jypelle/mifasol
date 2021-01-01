package mobilecli

import (
	"encoding/json"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/jypelle/mifasol/internal/mobilecli/config"
	"github.com/jypelle/mifasol/internal/version"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

type PageType int

const (
	SyncPageType PageType = iota
	SetupPageType
)

type MobileApp struct {
	config     config.ClientConfig
	restClient *restClientV1.RestClient

	app        fyne.App
	windowType PageType
	window     fyne.Window
	syncPage   *SyncPage
	setupPage  *SetupPage
}

func NewMobileApp(configDir string, debugMode bool) *MobileApp {
	logrus.Debugf("Creation of mifasol mobile client %s ...", version.AppVersion.String())

	mobApp := &MobileApp{
		config: config.ClientConfig{
			ConfigDir: configDir,
			DebugMode: debugMode,
		},
		app:        app.New(),
		windowType: SyncPageType,
	}

	mobApp.window = mobApp.app.NewWindow("Mifasol")
	//	mobApp.window.Resize(fyne.NewSize(800, 700))

	mobApp.window.SetContent(
		widget.NewVBox(
			widget.NewLabel("Mifasol"),
			widget.NewButton("Quit", func() { mobApp.app.Quit() }),
		),
	)

	mobApp.window.Show()

	mobApp.app.Settings().SetTheme(theme.LightTheme())

	// Check Configuration folder
	_, err := os.Stat(mobApp.config.ConfigDir)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Printf("Creation of config folder: %s", mobApp.config.ConfigDir)
			err = os.Mkdir(mobApp.config.ConfigDir, 0770)
			if err != nil {
				logrus.Fatalf("Unable to create config folder: %v\n", err)
			}
		} else {
			logrus.Fatalf("Unable to access config folder: %s", mobApp.config.ConfigDir)
		}
	}

	// Open configuration file
	var draftClientEditableConfig *config.ClientEditableConfig

	rawConfig, err := ioutil.ReadFile(mobApp.config.GetCompleteConfigFilename())
	if err == nil {
		// Interpret configuration file
		draftClientEditableConfig = &config.ClientEditableConfig{}
		err = json.Unmarshal(rawConfig, draftClientEditableConfig)
		if err != nil {
			logrus.Fatalf("Unable to interpret config file: %v\n", err)
		}
	}

	mobApp.config.ClientEditableConfig = config.NewClientEditableConfig(draftClientEditableConfig)

	mobApp.config.Save()

	mobApp.syncPage = NewSyncPage(mobApp)
	mobApp.setupPage = NewSetupPage(mobApp)

	logrus.Debugln("Mobile Client created")

	return mobApp

}

func (c *MobileApp) Run() {
	c.app.Run()
}
