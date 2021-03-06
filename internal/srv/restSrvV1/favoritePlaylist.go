package restSrvV1

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"net/http"
)

func (s *RestServer) readFavoritePlaylists(w http.ResponseWriter, r *http.Request) {
	s.log.Debugf("Read favorite playlists")

	favoritePlaylists, err := s.store.ReadFavoritePlaylists(nil, &restApiV1.FavoritePlaylistFilter{})
	if err != nil {
		s.log.Panicf("Unable to read favorite playlists: %v", err)
	}

	tool.WriteJsonResponse(w, favoritePlaylists)
}

func (s *RestServer) createFavoritePlaylist(w http.ResponseWriter, r *http.Request) {
	s.log.Debugf("Create favorite playlist")

	var favoritePlaylistMeta restApiV1.FavoritePlaylistMeta
	err := json.NewDecoder(r.Body).Decode(&favoritePlaylistMeta)
	if err != nil {
		s.log.Panicf("Unable to interpret data to create the favorite playlist: %v", err)
	}

	favoritePlaylist, err := s.store.CreateFavoritePlaylist(nil, &favoritePlaylistMeta, true)
	if err != nil {
		s.log.Panicf("Unable to create the favorite playlist: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	tool.WriteJsonResponse(w, favoritePlaylist)
}

func (s *RestServer) deleteFavoritePlaylist(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	userId := restApiV1.UserId(vars["userId"])
	playlistId := restApiV1.PlaylistId(vars["playlistId"])
	favoritePlaylistId := restApiV1.FavoritePlaylistId{UserId: userId, PlaylistId: playlistId}

	s.log.Debugf("Delete favorite playlist: %v", favoritePlaylistId)

	favoritePlaylist, err := s.store.DeleteFavoritePlaylist(nil, favoritePlaylistId)
	if err != nil {
		s.log.Panicf("Unable to delete favorite playlist: %v", err)
	}

	tool.WriteJsonResponse(w, favoritePlaylist)

}
