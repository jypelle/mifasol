package restSrvV1

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jypelle/mifasol/internal/srv/storeerror"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"net/http"
)

func (s *RestServer) readAlbums(w http.ResponseWriter, r *http.Request) {
	s.log.Debugf("Read albums")

	var albumFilter restApiV1.AlbumFilter
	err := json.NewDecoder(r.Body).Decode(&albumFilter)
	if err != nil {
		s.log.Panicf("Unable to interpret data to read the albums: %v", err)
	}

	albums, err := s.store.ReadAlbums(nil, &albumFilter)
	if err != nil {
		s.log.Panicf("Unable to read albums: %v", err)
	}

	tool.WriteJsonResponse(w, albums)
}

func (s *RestServer) readAlbum(w http.ResponseWriter, r *http.Request) {
	s.log.Debugf("Read album")

	vars := mux.Vars(r)
	albumId := restApiV1.AlbumId(vars["id"])

	s.log.Debugf("Read album: %s", albumId)

	album, err := s.store.ReadAlbum(nil, albumId)
	if err != nil {
		if err == storeerror.ErrNotFound {
			s.apiErrorCodeResponse(w, restApiV1.NotFoundErrorCode)
			return
		}
		s.log.Panicf("Unable to read album: %v", err)
	}

	tool.WriteJsonResponse(w, album)
}

func (s *RestServer) createAlbum(w http.ResponseWriter, r *http.Request) {
	s.log.Debugf("Create album")

	var albumMeta restApiV1.AlbumMeta
	err := json.NewDecoder(r.Body).Decode(&albumMeta)
	if err != nil {
		s.log.Panicf("Unable to interpret data to create the album: %v", err)
	}

	// Check credential
	// TODO

	album, err := s.store.CreateAlbum(nil, &albumMeta)
	if err != nil {
		s.log.Panicf("Unable to create the album: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	tool.WriteJsonResponse(w, album)
}

func (s *RestServer) updateAlbum(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	albumId := restApiV1.AlbumId(vars["id"])

	s.log.Debugf("Update album: %s", albumId)

	var albumMeta restApiV1.AlbumMeta
	err := json.NewDecoder(r.Body).Decode(&albumMeta)
	if err != nil {
		s.log.Panicf("Unable to interpret data to update the album: %v", err)
	}

	album, err := s.store.UpdateAlbum(nil, albumId, &albumMeta)
	if err != nil {
		s.log.Panicf("Unable to update the album: %v", err)
	}

	tool.WriteJsonResponse(w, album)

}

func (s *RestServer) deleteAlbum(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	albumId := restApiV1.AlbumId(vars["id"])

	s.log.Debugf("Delete album: %s", albumId)

	album, err := s.store.DeleteAlbum(nil, albumId)
	if err != nil {
		if err == storeerror.ErrDeleteAlbumWithSongs {
			s.apiErrorCodeResponse(w, restApiV1.DeleteAlbumWithSongsErrorCode)
			return
		}

		s.log.Panicf("Unable to delete album: %v", err)
	}

	tool.WriteJsonResponse(w, album)

}
