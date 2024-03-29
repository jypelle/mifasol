package webSrv

import (
	"embed"
	"github.com/gorilla/mux"
	"github.com/jypelle/mifasol/internal/srv/config"
	"github.com/jypelle/mifasol/internal/srv/store"
	"github.com/jypelle/mifasol/internal/srv/webSrv/clients"
	"github.com/jypelle/mifasol/internal/srv/webSrv/static"
	"github.com/jypelle/mifasol/internal/srv/webSrv/templates"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/internal/version"
	"github.com/sirupsen/logrus"
	"github.com/vearutop/statigz"
	"html/template"
	"io/fs"
	"net/http"
	"time"
)

type WebServer struct {
	store        *store.Store
	router       *mux.Router
	serverConfig *config.ServerConfig

	templateHelpers template.FuncMap

	StaticFs    embed.FS
	TemplatesFs fs.FS

	log *logrus.Entry
}

func NewWebServer(store *store.Store, router *mux.Router, serverConfig *config.ServerConfig) *WebServer {

	webServer := &WebServer{
		store:        store,
		router:       router,
		serverConfig: serverConfig,
		log:          logrus.WithField("origin", "web"),
	}

	// Ressources
	//	if serverConfig.EmbeddedFs {
	webServer.StaticFs = static.Fs
	webServer.TemplatesFs = templates.Fs
	//webServer.ClientsFs = clients.Fs
	//	} else {
	//		webServer.StaticFs = os.DirFS("internal/srv/webSrv/static")
	//		webServer.TemplatesFs = os.DirFS("internal/srv/webSrv/templates")
	//	}

	// Set routes

	// Static files
	//
	staticFileHandler := http.StripPrefix("/static", http.FileServer(
		http.FS(&tool.StaticFSWrapper{
			ReadDirFS:    webServer.StaticFs,
			FixedModTime: time.Now(),
		}),
	))

	webServer.router.PathPrefix("/static").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Expires", "")
		w.Header().Set("Cache-Control", "no-cache, max-age=0")
		w.Header().Set("Pragma", "")
		staticFileHandler.ServeHTTP(w, r)
	})

	// Clients binary executable files
	//
	clientsFileHandler := http.StripPrefix("/clients", statigz.FileServer(
		&tool.StaticFSWrapper{
			ReadDirFS:    clients.Fs,
			FixedModTime: time.Now(),
		},
	))
	webServer.router.PathPrefix("/clients").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Expires", "")
		w.Header().Set("Cache-Control", "no-cache, max-age=0")
		w.Header().Set("Pragma", "")
		clientsFileHandler.ServeHTTP(w, r)
	})

	// Start page
	webServer.router.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		webServer.HtmlWriterRender(w, "Mifasol", "main.html")
	}).Methods("GET").Name("start")

	// Service worker
	webServer.router.HandleFunc("/sw.js", func(w http.ResponseWriter, _ *http.Request) {
		webServer.JsWriterRender(w, version.AppVersion.String(), "sw.js")
	}).Methods("GET").Name("serviceWorker")

	return webServer
}

func (s *WebServer) Log() *logrus.Entry {
	return s.log
}

func (d *WebServer) Config() *config.ServerConfig {
	return d.serverConfig
}

type IndexView struct {
	Title string
}
