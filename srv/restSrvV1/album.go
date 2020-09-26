package restSrvV1

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/srv/svc"
	"github.com/jypelle/mifasol/tool"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (s *RestServer) readAlbums(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Read albums")

	albums, err := s.service.ReadAlbums(nil, &restApiV1.AlbumFilter{})
	if err != nil {
		logrus.Panicf("Unable to read albums: %v", err)
	}

	tool.WriteJsonResponse(w, albums)
}

func (s *RestServer) readAlbum(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Read album")

	vars := mux.Vars(r)
	albumId := restApiV1.AlbumId(vars["id"])

	logrus.Debugf("Read album: %s", albumId)

	album, err := s.service.ReadAlbum(nil, albumId)
	if err != nil {
		if err == svc.ErrNotFound {
			s.apiErrorCodeResponse(w, restApiV1.NotFoundErrorCode)
			return
		}
		logrus.Panicf("Unable to read album: %v", err)
	}

	tool.WriteJsonResponse(w, album)
}

func (s *RestServer) createAlbum(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Create album")

	var albumMeta restApiV1.AlbumMeta
	err := json.NewDecoder(r.Body).Decode(&albumMeta)
	if err != nil {
		logrus.Panicf("Unable to interpret data to create the album: %v", err)
	}

	// Check credential
	// TODO

	album, err := s.service.CreateAlbum(nil, &albumMeta)
	if err != nil {
		logrus.Panicf("Unable to create the album: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	tool.WriteJsonResponse(w, album)
}

func (s *RestServer) updateAlbum(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	albumId := restApiV1.AlbumId(vars["id"])

	logrus.Debugf("Update album: %s", albumId)

	var albumMeta restApiV1.AlbumMeta
	err := json.NewDecoder(r.Body).Decode(&albumMeta)
	if err != nil {
		logrus.Panicf("Unable to interpret data to update the album: %v", err)
	}

	album, err := s.service.UpdateAlbum(nil, albumId, &albumMeta)
	if err != nil {
		logrus.Panicf("Unable to update the album: %v", err)
	}

	tool.WriteJsonResponse(w, album)

}

func (s *RestServer) deleteAlbum(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	albumId := restApiV1.AlbumId(vars["id"])

	logrus.Debugf("Delete album: %s", albumId)

	album, err := s.service.DeleteAlbum(nil, albumId)
	if err != nil {
		if err == svc.ErrDeleteAlbumWithSongs {
			s.apiErrorCodeResponse(w, restApiV1.DeleteAlbumWithSongsErrorCode)
			return
		}

		logrus.Panicf("Unable to delete album: %v", err)
	}

	tool.WriteJsonResponse(w, album)

}
