package restSrvV1

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/jypelle/mifasol/internal/srv/svc"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"sync"
)

var contextKeyUser = "user"

type RestServer struct {
	service   *svc.Service
	subRooter *mux.Router

	sessionMap sync.Map
}

func NewRestServer(service *svc.Service, subRouter *mux.Router) *RestServer {

	restServer := &RestServer{
		service:   service,
		subRooter: subRouter,
	}

	restServer.subRooter.HandleFunc("/token", restServer.generateToken).Methods("POST")

	restServer.subRooter.HandleFunc("/albums", restServer.readAlbums).Methods("GET")
	restServer.subRooter.HandleFunc("/albums/{id}", restServer.readAlbum).Methods("GET")
	restServer.subRooter.HandleFunc("/albums", restServer.createAlbum).Methods("POST")        // Admin only
	restServer.subRooter.HandleFunc("/albums/{id}", restServer.updateAlbum).Methods("PUT")    // Admin only
	restServer.subRooter.HandleFunc("/albums/{id}", restServer.deleteAlbum).Methods("DELETE") // Admin only

	restServer.subRooter.HandleFunc("/artists", restServer.readArtists).Methods("GET")
	restServer.subRooter.HandleFunc("/artists/{id}", restServer.readArtist).Methods("GET")
	restServer.subRooter.HandleFunc("/artists", restServer.createArtist).Methods("POST")        // Admin only
	restServer.subRooter.HandleFunc("/artists/{id}", restServer.updateArtist).Methods("PUT")    // Admin only
	restServer.subRooter.HandleFunc("/artists/{id}", restServer.deleteArtist).Methods("DELETE") // Admin only

	restServer.subRooter.HandleFunc("/playlists", restServer.readPlaylists).Methods("GET")
	restServer.subRooter.HandleFunc("/playlists/{id}", restServer.readPlaylist).Methods("GET")
	restServer.subRooter.HandleFunc("/playlists", restServer.createPlaylist).Methods("POST")        // Admin only
	restServer.subRooter.HandleFunc("/playlists/{id}", restServer.updatePlaylist).Methods("PUT")    // Admin only
	restServer.subRooter.HandleFunc("/playlists/{id}", restServer.deletePlaylist).Methods("DELETE") // Admin only

	restServer.subRooter.HandleFunc("/songs", restServer.readSongs).Methods("GET")
	restServer.subRooter.HandleFunc("/songs/{id}", restServer.readSong).Methods("GET")
	restServer.subRooter.HandleFunc("/songContents/{id}", restServer.readSongContent).Methods("GET")
	restServer.subRooter.HandleFunc("/songContents", restServer.createSongContent).Methods("POST")                      // Admin only
	restServer.subRooter.HandleFunc("/songContentsForAlbum/{id}", restServer.createSongContentForAlbum).Methods("POST") // Admin only
	restServer.subRooter.HandleFunc("/songWithContents", restServer.createSongWithContent).Methods("POST")              // Admin only
	restServer.subRooter.HandleFunc("/songs/{id}", restServer.updateSong).Methods("PUT")                                // Admin only
	restServer.subRooter.HandleFunc("/songs/{id}", restServer.deleteSong).Methods("DELETE")                             // Admin only

	restServer.subRooter.HandleFunc("/users", restServer.readUsers).Methods("GET")
	restServer.subRooter.HandleFunc("/users/{id}", restServer.readUser).Methods("GET")
	restServer.subRooter.HandleFunc("/users", restServer.createUser).Methods("POST") // Admin only
	restServer.subRooter.HandleFunc("/users/{id}", restServer.updateUser).Methods("PUT")
	restServer.subRooter.HandleFunc("/users/{id}", restServer.deleteUser).Methods("DELETE") // Admin only

	restServer.subRooter.HandleFunc("/favoritePlaylists", restServer.readFavoritePlaylists).Methods("GET")
	restServer.subRooter.HandleFunc("/favoritePlaylists", restServer.createFavoritePlaylist).Methods("POST")
	restServer.subRooter.HandleFunc("/favoritePlaylists/{userId}/{playlistId}", restServer.deleteFavoritePlaylist).Methods("DELETE")

	restServer.subRooter.HandleFunc("/favoriteSongs", restServer.readFavoriteSongs).Methods("GET")
	restServer.subRooter.HandleFunc("/favoriteSongs", restServer.createFavoriteSong).Methods("POST")
	restServer.subRooter.HandleFunc("/favoriteSongs/{userId}/{songId}", restServer.deleteFavoriteSong).Methods("DELETE")

	restServer.subRooter.HandleFunc("/syncReport/{fromTs}", restServer.readSyncReport).Methods("GET")
	restServer.subRooter.HandleFunc("/fileSyncReport/{fromTs}/{userId}", restServer.readFileSyncReport).Methods("GET")

	restServer.subRooter.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		restServer.apiErrorCodeResponse(w, restApiV1.MethodNotAllowedErrorCode)
	})
	restServer.subRooter.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		restServer.apiErrorCodeResponse(w, restApiV1.NotFoundErrorCode)
	})

	restServer.subRooter.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logrus.Warningln("Recovering API Call...")
					restServer.apiErrorCodeResponse(w, restApiV1.InternalErrorCode)
				}
			}()

			// Check Token
			ctx := r.Context()
			if r.URL.Path != "/api/v1/token" {
				reqToken := r.Header.Get("Authorization")
				splitToken := strings.Split(reqToken, "Bearer")
				if len(splitToken) < 2 {
					restServer.apiErrorCodeResponse(w, restApiV1.InvalidTokenErrorCode)
					return
				}
				accessToken := strings.Trim(splitToken[1], " ")

				logrus.Debugln("Check token " + accessToken + " for " + r.URL.Path)

				ses, ok := restServer.sessionMap.Load(accessToken)
				if ok != true {
					restServer.apiErrorCodeResponse(w, restApiV1.InvalidTokenErrorCode)
					return
				}

				user, err := restServer.service.ReadUser(nil, ses.(*session).userId)
				if err != nil {
					if err == svc.ErrNotFound {
						restServer.apiErrorCodeResponse(w, restApiV1.InvalidTokenErrorCode)
						return
					}
					restServer.apiErrorCodeResponse(w, restApiV1.InternalErrorCode)
					return
				}
				logrus.Debugln("User: " + user.Name)
				ctx = context.WithValue(ctx, contextKeyUser, user)

			}

			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	return restServer
}

func (s *RestServer) GetConnectedUser(r *http.Request) *restApiV1.User {
	user, ok := r.Context().Value(contextKeyUser).(*restApiV1.User)
	if ok && user != nil {
		return user
	} else {
		return nil
	}
}

func (s *RestServer) IsAdmin(r *http.Request) bool {
	user := s.GetConnectedUser(r)
	return user != nil && user.AdminFg
}

func (s *RestServer) CheckAdmin(w http.ResponseWriter, r *http.Request) bool {
	if !s.IsAdmin(r) {
		s.apiErrorCodeResponse(w, restApiV1.ForbiddenErrorCode)
		return false
	} else {
		return true
	}
}
