package cliwa

import (
	"bytes"
	"github.com/jypelle/mifasol/internal/cliwa/config"
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/internal/cliwa/templates"
	"github.com/jypelle/mifasol/internal/localdb"
	"github.com/jypelle/mifasol/internal/version"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/sirupsen/logrus"
	"html/template"
	"net/url"
	"strconv"
	"syscall/js"
)

type App struct {
	config     config.ClientConfig
	restClient *restClientV1.RestClient
	localDb    *localdb.LocalDb

	templateHelpers template.FuncMap

	StartComponent *StartComponent
	HomeComponent  *HomeComponent

	eventFunc chan func()
}

func NewApp(debugMode bool) *App {

	logrus.Infof("Creation of mifasol web assembly client %s ...", version.AppVersion.String())

	app := &App{
		config: config.ClientConfig{
			ClientEditableConfig: config.NewClientEditableConfig(nil),
		},
		eventFunc: make(chan func(), 100),
	}

	app.StartComponent = NewStartComponent(app)

	logrus.Infof("Client created")

	return app
}

func (a *App) Start() {
	a.retrieveServerCredentials()
	a.HideLoader()
	a.StartComponent.Show()

	// Keep wasm app alive with event func loop
	func() {
		for {
			select {
			case f := <-a.eventFunc:
				f()
			}
		}
	}()
}

func (a *App) RenderTemplate(content interface{}, filenames ...string) string {

	// Parsing template files
	t := template.New("").Funcs(a.templateHelpers)

	for _, filename := range filenames {

		html, err := templates.Fs.ReadFile(filename + ".html")
		if err != nil {
			logrus.Panicf("Unable to read template file %s: %v\n", filename, err)
		}

		t, err = t.Parse(string(html))
		if err != nil {
			logrus.Panicf("Unable to interpret template file %s: %v\n", filename, err)
		}
	}

	var w bytes.Buffer
	err := t.Execute(&w, content)
	if err != nil {
		logrus.Panicf("Unable to execute template files : %v\n", err)
	}

	return w.String()
}

func (a *App) AddBlockingRichEventFunc(fn func(this js.Value, args []js.Value)) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fn(this, args)
		return nil
	})
}

func (a *App) AddRichEventFunc(fn func(this js.Value, args []js.Value)) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.eventFunc <- func() {
			fn(this, args)
		}
		return nil
	})
}

func (a *App) AddEventFunc(fn func()) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.eventFunc <- func() {
			fn()
		}
		return nil
	})
}

func (a *App) AddEventFuncPreventDefault(fn func()) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		args[0].Call("preventDefault")
		a.eventFunc <- func() {
			fn()
		}
		return nil
	})
}

func (a *App) retrieveServerCredentials() {
	rawUrl := js.Global().Get("window").Get("location").Get("href").String()
	baseUrl, _ := url.Parse(rawUrl)

	a.config.ServerHostname = baseUrl.Hostname()
	a.config.ServerPort, _ = strconv.ParseInt(baseUrl.Port(), 10, 64)
	a.config.ServerSsl = baseUrl.Scheme == "https"
	if a.config.ServerPort == 0 {
		if a.config.ServerSsl {
			a.config.ServerPort = 443
		} else {
			a.config.ServerPort = 80
		}
	}
}

func (a *App) ConnectedUserId() restApiV1.UserId {
	return a.restClient.UserId()
}

func (a *App) IsConnectedUserAdmin() bool {
	if user, ok := a.localDb.Users[a.ConnectedUserId()]; ok == true {
		return user.AdminFg
	}
	return false
}

func (a *App) HideExplicitSongForConnectedUser() bool {
	if user, ok := a.localDb.Users[a.ConnectedUserId()]; ok == true {
		return user.HideExplicitFg
	}
	return false
}

func (c *App) ShowDefaultLoader(message string) {
	c.ShowLoader("Loading")
}

func (c *App) ShowLoader(message string) {
	jst.Id("modalLoaderMessage").Set("innerHTML", message)
	jst.Document.Get("body").Get("classList").Call("add", "loading")
}

func (c *App) HideLoader() {
	jst.Document.Get("body").Get("classList").Call("remove", "loading")
}
