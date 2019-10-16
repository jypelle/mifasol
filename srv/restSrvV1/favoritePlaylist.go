package restSrvV1

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"mifasol/restApiV1"
	"mifasol/tool"
	"net/http"
)

func (s *RestServer) readFavoritePlaylists(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Read favorite playlists")

	favoritePlaylists, err := s.service.ReadFavoritePlaylists(nil, &restApiV1.FavoritePlaylistFilter{})
	if err != nil {
		logrus.Panicf("Unable to read favorite playlists: %v", err)
	}

	tool.WriteJsonResponse(w, favoritePlaylists)
}

func (s *RestServer) createFavoritePlaylist(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Create favorite playlist")

	var favoritePlaylistMeta restApiV1.FavoritePlaylistMeta
	err := json.NewDecoder(r.Body).Decode(&favoritePlaylistMeta)
	if err != nil {
		logrus.Panicf("Unable to interpret data to create the favorite playlist: %v", err)
	}

	favoritePlaylist, err := s.service.CreateFavoritePlaylist(nil, &favoritePlaylistMeta, true)
	if err != nil {
		logrus.Panicf("Unable to create the favorite playlist: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	tool.WriteJsonResponse(w, favoritePlaylist)
}

func (s *RestServer) deleteFavoritePlaylist(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	userId := vars["userId"]
	playlistId := vars["playlistId"]
	favoritePlaylistId := restApiV1.FavoritePlaylistId{UserId: userId, PlaylistId: playlistId}

	logrus.Debugf("Delete favorite playlist: %v", favoritePlaylistId)

	favoritePlaylist, err := s.service.DeleteFavoritePlaylist(nil, favoritePlaylistId)
	if err != nil {
		logrus.Panicf("Unable to delete favorite playlist: %v", err)
	}

	tool.WriteJsonResponse(w, favoritePlaylist)

}
