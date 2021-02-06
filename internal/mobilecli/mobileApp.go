package mobilecli

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/jypelle/mifasol/internal/mobilecli/config"
	"github.com/jypelle/mifasol/internal/version"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/sirupsen/logrus"
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
	mobApp.window.Resize(fyne.NewSize(800, 700))

	testButton := widget.NewButton("Yoko!", func() {})

	mobApp.window.SetContent(
		container.New(layout.NewMaxLayout(),
			container.NewAppTabs(
				container.NewTabItem(
					"Song",
					container.NewVScroll(
						container.New(layout.NewVBoxLayout(),
							widget.NewButton("aa", func() {}),
						),
					),
				),
				container.NewTabItem(
					"Artist",
					container.NewVScroll(
						container.New(layout.NewVBoxLayout(),
							widget.NewButton("bb", func() {}),
						),
					),
				),
				container.NewTabItem(
					"Album",
					container.NewVScroll(
						container.New(layout.NewVBoxLayout(),
							widget.NewButton("cc", func() {}),
						),
					),
				),
				container.NewTabItem(
					"Playlist",
					container.New(
						layout.NewBorderLayout(nil, testButton, nil, nil),
						container.NewVScroll(
							container.New(layout.NewVBoxLayout(),
								widget.NewLabel("Test1"),
								widget.NewLabel("Test2"),
								widget.NewLabel("Test3"),
								widget.NewLabel("Test4"),
								widget.NewLabel("Test5"),
								widget.NewLabel("Test6"),
								widget.NewLabel("Test7"),
								widget.NewLabel("Test8"),
								widget.NewLabel("Test9"),
								widget.NewLabel("Test10"),
								widget.NewLabel("Test11"),
								widget.NewLabel("Test12"),
								widget.NewLabel("Test13"),
								widget.NewLabel("Test14"),
								widget.NewLabel("Test15"),
								widget.NewLabel("Test16"),
								widget.NewLabel("Test17"),
								widget.NewLabel("Test18"),
								widget.NewLabel("Test19"),
								widget.NewLabel("Test20"),
								widget.NewLabel("Test1"),
								widget.NewLabel("Test2"),
								widget.NewLabel("Test3"),
								widget.NewLabel("Test4"),
								widget.NewLabel("Test5"),
								widget.NewLabel("Test6"),
								widget.NewLabel("Test7"),
								widget.NewLabel("Test8"),
								widget.NewLabel("Test9"),
								widget.NewLabel("Test10"),
								widget.NewLabel("Test11"),
								widget.NewLabel("Test12"),
								widget.NewLabel("Test13"),
								widget.NewLabel("Test14"),
								widget.NewLabel("Test15"),
								widget.NewLabel("Test16"),
								widget.NewLabel("Test17"),
								widget.NewLabel("Test18"),
								widget.NewLabel("Test19"),
								widget.NewLabel("Test20"),
								widget.NewLabel("Test1"),
								widget.NewLabel("Test2"),
								widget.NewLabel("Test3"),
								widget.NewLabel("Test4"),
								widget.NewLabel("Test5"),
								widget.NewLabel("Test6"),
								widget.NewLabel("Test7"),
								widget.NewLabel("Test8"),
								widget.NewLabel("Test9"),
								widget.NewLabel("Test10"),
								widget.NewLabel("Test11"),
								widget.NewLabel("Test12"),
								widget.NewLabel("Test13"),
								widget.NewLabel("Test14"),
								widget.NewLabel("Test15"),
								widget.NewLabel("Test16"),
								widget.NewLabel("Test17"),
								widget.NewLabel("Test18"),
								widget.NewLabel("Test19"),
								widget.NewLabel("Test20"),
							),
						),
						testButton,
					),
				),
				container.NewTabItem(
					"Config",
					container.NewVScroll(
						container.New(layout.NewVBoxLayout(),
							widget.NewButton("Sync", func() {}),
						),
					),
				),
			),
			//		widget.NewLabel("Mifasol"),
			//		widget.NewButton("Play", func() {}),
			//		widget.NewButton("Pause", func() {}),
		),
	)

	mobApp.window.Show()

	mobApp.app.Settings().SetTheme(theme.LightTheme())
	/*
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
	*/
	mobApp.syncPage = NewSyncPage(mobApp)
	mobApp.setupPage = NewSetupPage(mobApp)

	logrus.Debugln("Mobile Client created")

	return mobApp

}

func (c *MobileApp) Run() {
	c.app.Run()
}
