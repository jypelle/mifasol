package restSrvV1

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (s *RestServer) readFavoriteSongs(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Read favorite songs")

	favoriteSongs, err := s.oldStore.ReadFavoriteSongs(nil, &restApiV1.FavoriteSongFilter{})
	if err != nil {
		logrus.Panicf("Unable to read favorite songs: %v", err)
	}

	tool.WriteJsonResponse(w, favoriteSongs)
}

func (s *RestServer) createFavoriteSong(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Create favorite song")

	var favoriteSongMeta restApiV1.FavoriteSongMeta
	err := json.NewDecoder(r.Body).Decode(&favoriteSongMeta)
	if err != nil {
		logrus.Panicf("Unable to interpret data to create the favorite song: %v", err)
	}

	favoriteSong, err := s.oldStore.CreateFavoriteSong(nil, &favoriteSongMeta, true)
	if err != nil {
		logrus.Panicf("Unable to create the favorite song: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	tool.WriteJsonResponse(w, favoriteSong)
}

func (s *RestServer) deleteFavoriteSong(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	userId := restApiV1.UserId(vars["userId"])
	songId := restApiV1.SongId(vars["songId"])
	favoriteSongId := restApiV1.FavoriteSongId{UserId: userId, SongId: songId}

	logrus.Debugf("Delete favorite song: %v", favoriteSongId)

	favoriteSong, err := s.oldStore.DeleteFavoriteSong(nil, favoriteSongId)
	if err != nil {
		logrus.Panicf("Unable to delete favorite song: %v", err)
	}

	tool.WriteJsonResponse(w, favoriteSong)

}
