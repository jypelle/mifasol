package restSrvV1

/*
func (s *RestServer) readFavoritePlaylists(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Read favorite playlists")

	favoritePlaylists, err := s.service.ReadFavoritePlaylists(nil, &restApiV1.FavoritePlaylistFilter{})
	if err != nil {
		logrus.Panicf("Unable to read favorite playlists: %v", err)
	}

	tool.WriteJsonResponse(w, favoritePlaylists)
}


func (s *RestServer) createPlaylist(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Create playlist")

	var playlistMeta restApiV1.PlaylistMeta
	err := json.NewDecoder(r.Body).Decode(&playlistMeta)
	if err != nil {
		logrus.Panicf("Unable to interpret data to create the playlist: %v", err)
	}

	playlist, err := s.service.CreatePlaylist(nil, &playlistMeta, true)
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

	playlist, err := s.service.UpdatePlaylist(nil, playlistId, &playlistMeta, true)
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
*/
