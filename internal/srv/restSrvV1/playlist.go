package restSrvV1

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jypelle/mifasol/internal/srv/storeerror"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"net/http"
)

func (s *RestServer) readPlaylists(w http.ResponseWriter, r *http.Request) {
	s.log.Debugf("Read playlists")

	var playlistFilter restApiV1.PlaylistFilter
	err := json.NewDecoder(r.Body).Decode(&playlistFilter)
	if err != nil {
		s.log.Panicf("Unable to interpret data to read the playlists: %v", err)
	}

	playlists, err := s.store.ReadPlaylists(nil, &playlistFilter)
	if err != nil {
		s.log.Panicf("Unable to read playlists: %v", err)
	}

	tool.WriteJsonResponse(w, playlists)
}

func (s *RestServer) readPlaylist(w http.ResponseWriter, r *http.Request) {
	s.log.Debugf("Read playlist")

	vars := mux.Vars(r)
	playlistId := restApiV1.PlaylistId(vars["id"])

	s.log.Debugf("Read playlist: %s", playlistId)

	playlist, err := s.store.ReadPlaylist(nil, playlistId)
	if err != nil {
		if err == storeerror.ErrNotFound {
			s.apiErrorCodeResponse(w, restApiV1.NotFoundErrorCode)
			return
		}
		s.log.Panicf("Unable to read playlist: %v", err)
	}

	tool.WriteJsonResponse(w, playlist)
}

func (s *RestServer) createPlaylist(w http.ResponseWriter, r *http.Request) {
	s.log.Debugf("Create playlist")

	var playlistMeta restApiV1.PlaylistMeta
	err := json.NewDecoder(r.Body).Decode(&playlistMeta)
	if err != nil {
		s.log.Panicf("Unable to interpret data to create the playlist: %v", err)
	}

	playlist, err := s.store.CreatePlaylist(nil, &playlistMeta, true)
	if err != nil {
		s.log.Panicf("Unable to create the playlist: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	tool.WriteJsonResponse(w, playlist)
}

func (s *RestServer) updatePlaylist(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	playlistId := restApiV1.PlaylistId(vars["id"])

	s.log.Debugf("Update playlist: %s", playlistId)

	var playlistMeta restApiV1.PlaylistMeta
	err := json.NewDecoder(r.Body).Decode(&playlistMeta)
	if err != nil {
		s.log.Panicf("Unable to interpret data to update the playlist: %v", err)
	}

	playlist, err := s.store.UpdatePlaylist(nil, playlistId, &playlistMeta, true)
	if err != nil {
		s.log.Panicf("Unable to update the playlist: %v", err)
	}

	tool.WriteJsonResponse(w, playlist)

}

func (s *RestServer) deletePlaylist(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	playlistId := restApiV1.PlaylistId(vars["id"])

	s.log.Debugf("Delete playlist: %s", playlistId)

	playlist, err := s.store.DeletePlaylist(nil, playlistId)
	if err != nil {
		s.log.Panicf("Unable to delete playlist: %v", err)
	}

	tool.WriteJsonResponse(w, playlist)

}
