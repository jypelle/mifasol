package cliwa

import (
	"fmt"
	"github.com/jypelle/mifasol/internal/cli/config"
	"github.com/jypelle/mifasol/internal/version"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/sirupsen/logrus"
	"syscall/js"
	"time"
)

type App struct {
	config     config.ClientConfig
	restClient *restClientV1.RestClient
}

func NewApp(debugMode bool) *App {

	logrus.Infof("Creation of mifasol web assembly client %s ...", version.AppVersion.String())

	app := &App{
		//		config: config.ClientConfig{
		//			ConfigDir: configDir,
		//			DebugMode: debugMode,
		//		},
	}

	logrus.Infof("Client created")

	return app
}

func (c *App) Start() {
	/*
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

	*/

	fmt.Println("coucou")
	doc := js.Global().Get("document")
	body := doc.Get("body")
	body.Set("innerHTML", `<main>
	    <h1>Welcome to mifasol</h1>
	    <h2>Download your client</h2>
	    <ul>
	        <li><a href="/clients/mifasolcli-windows-amd64.exe">Windows (amd64)</a></li>
	        <li><a href="/clients/mifasolcli-linux-amd64">Linux (amd64)</a></li>
	        <li><a href="/clients/mifasolcli-linux-arm">Linux (arm)</a></li>
	    </ul>
		<p id="message">...aa</p>
	</main>`)

	message := doc.Call("getElementById", "message")
	time.Sleep(2 * time.Second)
	message.Set("innerHTML", "5")
	time.Sleep(2 * time.Second)
	message.Set("innerHTML", "4")
	time.Sleep(2 * time.Second)
	message.Set("innerHTML", "3")
	time.Sleep(2 * time.Second)
	message.Set("innerHTML", "2")
	time.Sleep(2 * time.Second)
	message.Set("innerHTML", "1")
	time.Sleep(2 * time.Second)
	message.Set("innerHTML", "0")

}
