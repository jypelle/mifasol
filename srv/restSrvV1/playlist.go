package restSrvV1

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"lyra/restApiV1"
	"lyra/srv/svc"
	"lyra/tool"
	"net/http"
)

func (s *RestServer) readPlaylists(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Read playlists")

	playlists, err := s.service.ReadPlaylists(nil, &restApiV1.PlaylistFilter{Order: restApiV1.PlaylistOrderByPlaylistName})
	if err != nil {
		logrus.Panicf("Unable to read playlists: %v", err)
	}

	tool.WriteJsonResponse(w, playlists)
}

func (s *RestServer) readPlaylist(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Read playlist")

	vars := mux.Vars(r)
	playlistId := vars["id"]

	logrus.Debugf("Read playlist: %s", playlistId)

	playlist, err := s.service.ReadPlaylist(nil, playlistId)
	if err != nil {
		if err == svc.ErrNotFound {
			s.apiErrorCodeResponse(w, restApiV1.NotFoundErrorCode)
			return
		}
		logrus.Panicf("Unable to read playlist: %v", err)
	}

	tool.WriteJsonResponse(w, playlist)
}

func (s *RestServer) createPlaylist(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Create playlist")

	var playlistMeta restApiV1.PlaylistMeta
	err := json.NewDecoder(r.Body).Decode(&playlistMeta)
	if err != nil {
		logrus.Panicf("Unable to interpret data to create the playlist: %v", err)
	}

	playlist, err := s.service.CreatePlaylist(nil, &playlistMeta)
	if err != nil {
		logrus.Panicf("Unable to create the playlist: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	tool.WriteJsonResponse(w, playlist)
}

func (s *RestServer) updatePlaylist(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	playlistId := vars["id"]

	logrus.Debugf("Update playlist: %s", playlistId)

	var playlistMeta restApiV1.PlaylistMeta
	err := json.NewDecoder(r.Body).Decode(&playlistMeta)
	if err != nil {
		logrus.Panicf("Unable to interpret data to update the playlist: %v", err)
	}

	playlist, err := s.service.UpdatePlaylist(nil, playlistId, &playlistMeta)
	if err != nil {
		logrus.Panicf("Unable to update the playlist: %v", err)
	}

	tool.WriteJsonResponse(w, playlist)

}

func (s *RestServer) deletePlaylist(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	playlistId := vars["id"]

	logrus.Debugf("Delete playlist: %s", playlistId)

	playlist, err := s.service.DeletePlaylist(nil, playlistId)
	if err != nil {
		logrus.Panicf("Unable to delete playlist: %v", err)
	}

	tool.WriteJsonResponse(w, playlist)

}
