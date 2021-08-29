package restSrvV1

import (
	"github.com/gorilla/mux"
	"github.com/jypelle/mifasol/internal/srv/store"
	"github.com/jypelle/mifasol/internal/srv/storeerror"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"sync"
)

type RestServer struct {
	store     *store.Store
	subRouter *mux.Router

	sessionMap sync.Map

	log *logrus.Entry
}

func NewRestServer(store *store.Store, subRouter *mux.Router) *RestServer {

	restServer := &RestServer{
		store:     store,
		subRouter: subRouter,
		log:       logrus.WithField("origin", "rest"),
	}

	restServer.subRouter.HandleFunc("/token", restServer.generateToken).Methods("POST")

	restServer.subRouter.HandleFunc("/albums", restServer.readAlbums).Methods("GET")
	restServer.subRouter.HandleFunc("/albums", restServer.readAlbums).Methods("POST").Headers("x-http-method-override", "GET")
	restServer.subRouter.HandleFunc("/albums/{id}", restServer.readAlbum).Methods("GET")
	restServer.subRouter.HandleFunc("/albums", restServer.createAlbum).Methods("POST")
	restServer.subRouter.HandleFunc("/albums/{id}", restServer.updateAlbum).Methods("PUT")
	restServer.subRouter.HandleFunc("/albums/{id}", restServer.deleteAlbum).Methods("DELETE")

	restServer.subRouter.HandleFunc("/artists", restServer.readArtists).Methods("GET")
	restServer.subRouter.HandleFunc("/artists", restServer.readArtists).Methods("POST").Headers("x-http-method-override", "GET")
	restServer.subRouter.HandleFunc("/artists/{id}", restServer.readArtist).Methods("GET")
	restServer.subRouter.HandleFunc("/artists", restServer.createArtist).Methods("POST")
	restServer.subRouter.HandleFunc("/artists/{id}", restServer.updateArtist).Methods("PUT")
	restServer.subRouter.HandleFunc("/artists/{id}", restServer.deleteArtist).Methods("DELETE")

	restServer.subRouter.HandleFunc("/playlists", restServer.readPlaylists).Methods("GET")
	restServer.subRouter.HandleFunc("/playlists", restServer.readPlaylists).Methods("POST").Headers("x-http-method-override", "GET")
	restServer.subRouter.HandleFunc("/playlists/{id}", restServer.readPlaylist).Methods("GET")
	restServer.subRouter.HandleFunc("/playlists", restServer.createPlaylist).Methods("POST")
	restServer.subRouter.HandleFunc("/playlists/{id}", restServer.updatePlaylist).Methods("PUT")
	restServer.subRouter.HandleFunc("/playlists/{id}", restServer.deletePlaylist).Methods("DELETE")

	restServer.subRouter.HandleFunc("/songs", restServer.readSongs).Methods("GET")
	restServer.subRouter.HandleFunc("/songs", restServer.readSongs).Methods("POST").Headers("x-http-method-override", "GET")
	restServer.subRouter.HandleFunc("/songs/{id}", restServer.readSong).Methods("GET")
	restServer.subRouter.HandleFunc("/songContents/{id}", restServer.readSongContent).Methods("GET")
	restServer.subRouter.HandleFunc("/songContents", restServer.createSongContent).Methods("POST")
	restServer.subRouter.HandleFunc("/songContentsForAlbum/{id}", restServer.createSongContentForAlbum).Methods("POST")
	restServer.subRouter.HandleFunc("/songWithContents", restServer.createSongWithContent).Methods("POST")
	restServer.subRouter.HandleFunc("/songs/{id}", restServer.updateSong).Methods("PUT")
	restServer.subRouter.HandleFunc("/songs/{id}", restServer.deleteSong).Methods("DELETE")

	restServer.subRouter.HandleFunc("/users", restServer.readUsers).Methods("GET")
	restServer.subRouter.HandleFunc("/users", restServer.readUsers).Methods("POST").Headers("x-http-method-override", "GET")
	restServer.subRouter.HandleFunc("/users/{id}", restServer.readUser).Methods("GET")
	restServer.subRouter.HandleFunc("/users", restServer.createUser).Methods("POST")
	restServer.subRouter.HandleFunc("/users/{id}", restServer.updateUser).Methods("PUT")
	restServer.subRouter.HandleFunc("/users/{id}", restServer.deleteUser).Methods("DELETE")

	restServer.subRouter.HandleFunc("/favoritePlaylists", restServer.readFavoritePlaylists).Methods("GET")
	restServer.subRouter.HandleFunc("/favoritePlaylists", restServer.readFavoritePlaylists).Methods("POST").Headers("x-http-method-override", "GET")
	restServer.subRouter.HandleFunc("/favoritePlaylists", restServer.createFavoritePlaylist).Methods("POST")
	restServer.subRouter.HandleFunc("/favoritePlaylists/{userId}/{playlistId}", restServer.deleteFavoritePlaylist).Methods("DELETE")

	restServer.subRouter.HandleFunc("/favoriteSongs", restServer.readFavoriteSongs).Methods("GET")
	restServer.subRouter.HandleFunc("/favoriteSongs", restServer.readFavoriteSongs).Methods("POST").Headers("x-http-method-override", "GET")
	restServer.subRouter.HandleFunc("/favoriteSongs", restServer.createFavoriteSong).Methods("POST")
	restServer.subRouter.HandleFunc("/favoriteSongs/{userId}/{songId}", restServer.deleteFavoriteSong).Methods("DELETE")

	restServer.subRouter.HandleFunc("/syncReport/{fromTs}", restServer.readSyncReport).Methods("GET")
	restServer.subRouter.HandleFunc("/fileSyncReport/{fromTs}/{userId}", restServer.readFileSyncReport).Methods("GET")

	restServer.subRouter.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		restServer.apiErrorCodeResponse(w, restApiV1.MethodNotAllowedErrorCode)
	})
	restServer.subRouter.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		restServer.apiErrorCodeResponse(w, restApiV1.NotFoundErrorCode)
	})

	restServer.subRouter.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					restServer.log.Warningln("Recovering API Call...")
					restServer.apiErrorCodeResponse(w, restApiV1.InternalErrorCode)
				}
			}()

			// Check Token
			if r.URL.Path != "/api/v1/token" {
				var accessToken string

				reqToken := r.Header.Get("Authorization")
				if reqToken != "" {
					splitToken := strings.Split(reqToken, "Bearer")
					if len(splitToken) == 2 {
						accessToken = strings.Trim(splitToken[1], " ")
					}
				} else {
					reqTokens, ok := r.URL.Query()["bearer"]
					if ok || len(reqTokens) == 1 {
						accessToken = reqTokens[0]
					}
				}

				if accessToken == "" {
					restServer.apiErrorCodeResponse(w, restApiV1.InvalidTokenErrorCode)
					return
				}

				restServer.log.Debugln("Check token " + accessToken + " for " + r.URL.Path)

				ses, ok := restServer.sessionMap.Load(accessToken)
				if ok != true {
					restServer.apiErrorCodeResponse(w, restApiV1.InvalidTokenErrorCode)
					return
				}

				user, err := restServer.store.ReadUser(nil, ses.(*session).userId)
				if err != nil {
					if err == storeerror.ErrNotFound {
						restServer.apiErrorCodeResponse(w, restApiV1.InvalidTokenErrorCode)
						return
					}
					restServer.apiErrorCodeResponse(w, restApiV1.InternalErrorCode)
					return
				}
				restServer.log.Debugln("User: " + user.Name)
				//			context.WithValue(r.Context(), contextKeyUser, user)

			}

			handler.ServeHTTP(w, r)
		})
	})

	return restServer
}
