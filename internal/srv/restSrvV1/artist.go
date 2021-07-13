package restSrvV1

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jypelle/mifasol/internal/srv/storeerror"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"net/http"
)

func (s *RestServer) readArtists(w http.ResponseWriter, r *http.Request) {
	s.log.Debugf("Read artists")

	var artistFilter restApiV1.ArtistFilter
	err := json.NewDecoder(r.Body).Decode(&artistFilter)
	if err != nil {
		s.log.Panicf("Unable to interpret data to read the artists: %v", err)
	}

	artists, err := s.store.ReadArtists(nil, &artistFilter)
	if err != nil {
		s.log.Panicf("Unable to read artists: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	tool.WriteJsonResponse(w, artists)
}

func (s *RestServer) readArtist(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	artistId := restApiV1.ArtistId(vars["id"])

	s.log.Debugf("Read artist: %s", artistId)

	artist, err := s.store.ReadArtist(nil, artistId)
	if err != nil {
		if err == storeerror.ErrNotFound {
			s.apiErrorCodeResponse(w, restApiV1.NotFoundErrorCode)
			return
		}
		s.log.Panicf("Unable to read artist: %v", err)
	}

	tool.WriteJsonResponse(w, artist)
}

func (s *RestServer) createArtist(w http.ResponseWriter, r *http.Request) {
	s.log.Debugf("Create artist")

	var artistMeta restApiV1.ArtistMeta
	err := json.NewDecoder(r.Body).Decode(&artistMeta)
	if err != nil {
		s.log.Panicf("Unable to interpret data to create the artist: %v", err)
	}

	artist, err := s.store.CreateArtist(nil, &artistMeta)
	if err != nil {
		s.log.Panicf("Unable to create the artist: %v", err)
	}

	tool.WriteJsonResponse(w, artist)
}

func (s *RestServer) updateArtist(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	artistId := restApiV1.ArtistId(vars["id"])

	s.log.Debugf("Update artist: %s", artistId)

	var artistMeta restApiV1.ArtistMeta
	err := json.NewDecoder(r.Body).Decode(&artistMeta)
	if err != nil {
		s.log.Panicf("Unable to interpret data to update the artist: %v", err)
	}

	artist, err := s.store.UpdateArtist(nil, artistId, &artistMeta)
	if err != nil {
		s.log.Panicf("Unable to update the artist: %v", err)
	}

	tool.WriteJsonResponse(w, artist)

}

func (s *RestServer) deleteArtist(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	artistId := restApiV1.ArtistId(vars["id"])

	s.log.Debugf("Delete artist: %s", artistId)

	artist, err := s.store.DeleteArtist(nil, artistId)
	if err != nil {
		if err == storeerror.ErrDeleteArtistWithSongs {
			s.apiErrorCodeResponse(w, restApiV1.DeleteArtistWithSongsErrorCode)
			return
		}

		s.log.Panicf("Unable to delete artist: %v", err)
	}

	tool.WriteJsonResponse(w, artist)

}
