package srv

import (
	"context"
	"encoding/json"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jypelle/mifasol/internal/srv/config"
	"github.com/jypelle/mifasol/internal/srv/oldstore"
	"github.com/jypelle/mifasol/internal/srv/restSrvV1"
	"github.com/jypelle/mifasol/internal/srv/store"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/internal/version"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	_ "modernc.org/sqlite"
	"net/http"
	"os"
	"strconv"
	"time"
)

type ServerApp struct {
	config.ServerConfig
	store      *store.Store
	restApiV1  *restSrvV1.RestServer
	httpServer *http.Server
}

func NewServerApp(configDir string, debugMode bool) *ServerApp {

	logrus.Debugf("Creation of mifasol server %s ...", version.AppVersion.String())

	app := &ServerApp{
		ServerConfig: config.ServerConfig{
			ConfigDir: configDir,
			DebugMode: debugMode,
		},
	}

	// Check Configuration folder
	_, err := os.Stat(app.ConfigDir)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Printf("Creation of config folder: %s", app.ConfigDir)
			err = os.Mkdir(app.ConfigDir, 0770)
			if err != nil {
				logrus.Fatalf("Unable to create config folder: %v\n", err)
			}
			logrus.Printf("Creation of songs folder: %s", app.ServerConfig.GetCompleteConfigSongsDirName())
			os.MkdirAll(app.ServerConfig.GetCompleteConfigSongsDirName(), 0770)
			logrus.Printf("Creation of albums folder: %s", app.ServerConfig.GetCompleteConfigAlbumsDirName())
			os.MkdirAll(app.ServerConfig.GetCompleteConfigAlbumsDirName(), 0770)
			logrus.Printf("Creation of authors folder: %s", app.ServerConfig.GetCompleteConfigAuthorsDirName())
			os.MkdirAll(app.ServerConfig.GetCompleteConfigAuthorsDirName(), 0770)

		} else {
			logrus.Fatalf("Unable to access config folder: %s", app.ConfigDir)
		}
	}

	// Open configuration file
	var draftServerEditableConfig *config.ServerEditableConfig

	rawConfig, err := ioutil.ReadFile(app.ServerConfig.GetCompleteConfigFilename())
	if err == nil {
		// Interpret configuration file
		draftServerEditableConfig = &config.ServerEditableConfig{}
		err = json.Unmarshal(rawConfig, draftServerEditableConfig)
		if err != nil {
			logrus.Fatalf("Unable to interpret config file: %v\n", err)
		}
	}

	app.ServerEditableConfig = config.NewServerEditableConfig(draftServerEditableConfig)

	app.ServerConfig.Save()

	if app.Ssl {
		existServerCert, err := tool.IsFileExists(app.GetCompleteConfigCertFilename())
		if err != nil {
			logrus.Fatalf("Unable to access %s: %v\n", app.GetCompleteConfigCertFilename(), err)
		}
		existServerKey, err := tool.IsFileExists(app.GetCompleteConfigKeyFilename())
		if err != nil {
			logrus.Fatalf("Unable to access %s: %v\n", app.GetCompleteConfigKeyFilename(), err)
		}

		if !existServerCert || !existServerKey {
			logrus.Info("Missing cert and key files, trying to generate them...")
			err = tool.GenerateTlsCertificate(
				"Mifasol",
				"Mifasol Server",
				app.GetCompleteConfigKeyFilename(),
				app.GetCompleteConfigCertFilename(),
				app.Hostnames)
			if err != nil {
				logrus.Fatalf("Unable to generate cert and key files : %v\n", err)
			}
			logrus.Info("Self-signed cert and key files generated")
		}

	}

	// Create store
	app.store = store.NewStore(&app.ServerConfig)

	// Migrate old store
	if _, err := os.Stat(app.ServerConfig.GetCompleteConfigOldDbFilename()); err == nil {
		logrus.Info("Migration of old key-value database...")
		oldStore := oldstore.NewOldStore(&app.ServerConfig)
		err = app.store.OldStoreMigration(oldStore)
		if err != nil {
			logrus.Fatalf("Unable to migrate old store : %v\n", err)
		}
		err := oldStore.Close()
		if err != nil {
			logrus.Fatalf("Unable to close the old store: %v", err)
		}
	}

	// Create router
	rooter := mux.NewRouter()

	// Create REST API
	app.restApiV1 = restSrvV1.NewRestServer(app.store, rooter.PathPrefix("/api/v1").Subrouter())

	// Create server check endpoint
	rooter.HandleFunc("/isalive",
		func(w http.ResponseWriter, r *http.Request) {
			logrus.Debugf("I'm alive")
			tool.WriteJsonResponse(w, true)
		}).Methods("GET")

	// Tell the browser that it's OK for JS to communicate with the server
	headersOk := handlers.AllowedHeaders([]string{"Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	app.httpServer = &http.Server{
		Addr:        ":" + strconv.FormatInt(app.Port, 10),
		Handler:     handlers.CORS(originsOk, headersOk, methodsOk)(app.recoverHandler(rooter)),
		ReadTimeout: time.Duration(app.Timeout) * time.Second,
	}

	logrus.Debugln("Server created")

	return app
}

func (s *ServerApp) Start() {
	logrus.Printf("Starting mifasol server ...")

	// Start serving REST request
	if s.Ssl {
		logrus.Printf("Server listening on https://localhost" + s.httpServer.Addr + " using a self-signed certificate")
		go func() {
			err := s.httpServer.ListenAndServeTLS(s.GetCompleteConfigCertFilename(), s.GetCompleteConfigKeyFilename())
			if err != nil && err != http.ErrServerClosed {
				logrus.Fatalf("Unable start the server: %v", err)
			}
		}()

	} else {
		logrus.Printf("Server listening on http://localhost" + s.httpServer.Addr)
		go func() {
			err := s.httpServer.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				logrus.Fatalf("Unable start the server: %v", err)
			}
		}()
	}

}

func (s *ServerApp) Stop() {
	logrus.Printf("Stopping mifasol server ...")

	// Stop listening REST request
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.httpServer.Shutdown(ctx)

	// Close store
	err := s.store.Close()
	if err != nil {
		logrus.Fatalf("Unable to close the database: %v", err)
	}

	logrus.Printf("Server stopped")
}

func (s *ServerApp) recoverHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec != nil {
				logrus.Warningln("Recovering...")

				http.Error(w, "Internal error", http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func (s *ServerApp) SaveConfig() {
	logrus.Debugf("Save config file: %s", s.GetCompleteConfigFilename())
	rawConfig, err := json.MarshalIndent(s.ServerConfig.ServerEditableConfig, "", "\t")
	if err != nil {
		logrus.Fatalf("Unable to serialize config file: %v\n", err)
	}
	err = ioutil.WriteFile(s.GetCompleteConfigFilename(), rawConfig, 0660)
	if err != nil {
		logrus.Fatalf("Unable to save config file: %v\n", err)
	}
}
