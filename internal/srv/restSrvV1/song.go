package restSrvV1

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jypelle/mifasol/internal/srv/svc"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func (s *RestServer) readSongs(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Read songs")

	songs, err := s.service.ReadSongs(nil, &restApiV1.SongFilter{})
	if err != nil {
		logrus.Panicf("Unable to read songs: %v", err)
	}

	tool.WriteJsonResponse(w, songs)
}

func (s *RestServer) readSong(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	songId := restApiV1.SongId(vars["id"])

	logrus.Debugf("Read song: %s", songId)

	song, err := s.service.ReadSong(nil, songId)
	if err != nil {
		if err == svc.ErrNotFound {
			s.apiErrorCodeResponse(w, restApiV1.NotFoundErrorCode)
			return
		}
		logrus.Panicf("Unable to read song: %v", err)
	}

	tool.WriteJsonResponse(w, song)
}

func (s *RestServer) readSongContent(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	songId := restApiV1.SongId(vars["id"])

	logrus.Debugf("Read song content: %s", songId)

	song, err := s.service.ReadSong(nil, songId)
	if err != nil {
		if err == svc.ErrNotFound {
			s.apiErrorCodeResponse(w, restApiV1.NotFoundErrorCode)
			return
		}
		logrus.Panicf("Unable to read song: %v", err)
	}

	songContent, err := s.service.ReadSongContent(song)
	if err != nil {
		if err == svc.ErrNotFound {
			s.apiErrorCodeResponse(w, restApiV1.NotFoundErrorCode)
			return
		}
		logrus.Panicf("Unable to read song content: %v", err)
	}

	w.Header().Set("Content-Type", song.Format.MimeType())
	w.Header().Set("Content-Length", strconv.Itoa(len(songContent)))

	w.Write(songContent)
}

func (s *RestServer) createSongContent(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Create song from raw content")

	song, err := s.service.CreateSongFromRawContent(nil, r.Body, restApiV1.UnknownAlbumId)

	if err != nil {
		logrus.Panicf("Unable to create the song: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	tool.WriteJsonResponse(w, song)
}

func (s *RestServer) createSongContentForAlbum(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Create song from raw content and try to link it to a specific albumId")

	vars := mux.Vars(r)
	lastAlbumId := restApiV1.AlbumId(vars["id"])

	song, err := s.service.CreateSongFromRawContent(nil, r.Body, lastAlbumId)

	if err != nil {
		logrus.Panicf("Unable to create the song: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	tool.WriteJsonResponse(w, song)
}

func (s *RestServer) createSongWithContent(w http.ResponseWriter, r *http.Request) {
	s.apiErrorCodeResponse(w, restApiV1.NotImplementedErrorCode)
	return
}

func (s *RestServer) updateSong(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	songId := restApiV1.SongId(vars["id"])

	logrus.Debugf("Update song: %s", songId)

	// Only admin
	if !s.CheckAdmin(w, r) {
		return
	}

	var songMeta restApiV1.SongMeta
	err := json.NewDecoder(r.Body).Decode(&songMeta)
	if err != nil {
		logrus.Panicf("Unable to interpret data to update the song: %v", err)
	}

	song, err := s.service.UpdateSong(nil, songId, &songMeta, nil, true)
	if err != nil {
		logrus.Panicf("Unable to update the song: %v", err)
	}

	tool.WriteJsonResponse(w, song)

}

func (s *RestServer) deleteSong(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	songId := restApiV1.SongId(vars["id"])

	logrus.Debugf("Delete song: %s", songId)

	// Only admin
	if !s.CheckAdmin(w, r) {
		return
	}

	song, err := s.service.DeleteSong(nil, songId)
	if err != nil {
		logrus.Panicf("Unable to delete song: %v", err)
	}

	tool.WriteJsonResponse(w, song)

}
