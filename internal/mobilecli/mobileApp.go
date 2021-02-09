package mobilecli

import (
	"encoding/json"
	"gioui.org/app"
	_ "gioui.org/app/permission/storage"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/jypelle/mifasol/internal/mobilecli/config"
	"github.com/jypelle/mifasol/internal/version"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/sirupsen/logrus"
	"image/color"
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

	th         *material.Theme
	windowType PageType
	window     *app.Window
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
		th:         material.NewTheme(gofont.Collection()),
		windowType: SyncPageType,
		window: app.NewWindow(
			app.Size(unit.Dp(800), unit.Dp(700)),
			app.Title("Mifasol"),
		),
	}

	mobApp.th.Palette.ContrastBg = color.NRGBA{
		R: 235,
		G: 70,
		B: 0,
		A: 255,
	}

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

	go func() {
		if err := c.loop(); err != nil {
			logrus.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()

}

func (c *MobileApp) loop() error {
	var ops op.Ops

	for {
		select {
		case e := <-c.window.Events():
			switch e := e.(type) {
			case *system.CommandEvent:
				switch c.windowType {
				case SyncPageType:

				case SetupPageType:
					c.setupPage.back()
					e.Cancel = true
				}

			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				switch c.windowType {
				case SyncPageType:
					c.syncPage.display(gtx, e.Insets)
				case SetupPageType:
					c.setupPage.display(gtx, e.Insets)
				}
				e.Frame(gtx.Ops)
			}
		}
	}
}
