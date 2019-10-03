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

func (s *RestServer) readArtists(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Read artists")

	artists, err := s.service.ReadArtists(nil, &restApiV1.ArtistFilter{Order: restApiV1.ArtistOrderByArtistName})
	if err != nil {
		logrus.Panicf("Unable to read artists: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	tool.WriteJsonResponse(w, artists)
}

func (s *RestServer) readArtist(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	artistId := vars["id"]

	logrus.Debugf("Read artist: %s", artistId)

	artist, err := s.service.ReadArtist(nil, artistId)
	if err != nil {
		if err == svc.ErrNotFound {
			s.apiErrorCodeResponse(w, restApiV1.NotFoundErrorCode)
			return
		}
		logrus.Panicf("Unable to read artist: %v", err)
	}

	tool.WriteJsonResponse(w, artist)
}

func (s *RestServer) createArtist(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Create artist")

	var artistMeta restApiV1.ArtistMeta
	err := json.NewDecoder(r.Body).Decode(&artistMeta)
	if err != nil {
		logrus.Panicf("Unable to interpret data to create the artist: %v", err)
	}

	artist, err := s.service.CreateArtist(nil, &artistMeta)
	if err != nil {
		logrus.Panicf("Unable to create the artist: %v", err)
	}

	tool.WriteJsonResponse(w, artist)
}

func (s *RestServer) updateArtist(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	artistId := vars["id"]

	logrus.Debugf("Update artist: %s", artistId)

	var artistMeta restApiV1.ArtistMeta
	err := json.NewDecoder(r.Body).Decode(&artistMeta)
	if err != nil {
		logrus.Panicf("Unable to interpret data to update the artist: %v", err)
	}

	artist, err := s.service.UpdateArtist(nil, artistId, &artistMeta)
	if err != nil {
		logrus.Panicf("Unable to update the artist: %v", err)
	}

	tool.WriteJsonResponse(w, artist)

}

func (s *RestServer) deleteArtist(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	artistId := vars["id"]

	logrus.Debugf("Delete artist: %s", artistId)

	artist, err := s.service.DeleteArtist(nil, artistId)
	if err != nil {
		if err == svc.ErrDeleteArtistWithSongs {
			s.apiErrorCodeResponse(w, restApiV1.DeleteArtistWithSongsErrorCode)
			return
		}

		logrus.Panicf("Unable to delete artist: %v", err)
	}

	tool.WriteJsonResponse(w, artist)

}
