package webSrv

import (
	"github.com/gorilla/mux"
	"github.com/jypelle/mifasol/internal/srv/config"
	"github.com/jypelle/mifasol/internal/srv/store"
	"github.com/jypelle/mifasol/internal/srv/webSrv/controller/home"
	"github.com/jypelle/mifasol/internal/srv/webSrv/static"
	"github.com/jypelle/mifasol/internal/srv/webSrv/templates"
	"github.com/sirupsen/logrus"
	"html/template"
	"io/fs"
	"net/http"
)

type WebServer struct {
	store        *store.Store
	router       *mux.Router
	serverConfig *config.ServerConfig

	templateHelpers template.FuncMap

	StaticFs    fs.FS
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
	//	} else {
	//		webServer.StaticFs = os.DirFS("internal/srv/webSrv/static")
	//		webServer.TemplatesFs = os.DirFS("internal/srv/webSrv/templates")
	//	}

	// Helpers & Components
	webServer.templateHelpers = template.FuncMap{
		"urlPath":      webServer.UrlPathHelper,
		"urlPathQuery": webServer.UrlPathQueryHelper,
		"queryParam":   webServer.QueryParamHelper,
		"partial":      webServer.PartialHelper,
	}

	// Set routes

	// Static files
	//
	staticFileHandler := http.StripPrefix("/static", http.FileServer(http.FS(webServer.StaticFs)))
	webServer.router.PathPrefix("/static").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Expires", "")
		w.Header().Set("Cache-Control", "public, max-age=2592000") // 30 days
		w.Header().Set("Pragma", "")
		staticFileHandler.ServeHTTP(w, r)
	})

	// Home
	homeController := home.NewController(webServer)
	webServer.router.HandleFunc("/", homeController.IndexAction).Methods("GET").Name("home")

	return webServer
}

func (s *WebServer) Log() *logrus.Entry {
	return s.log
}

func (d *WebServer) Config() *config.ServerConfig {
	return d.serverConfig
}
