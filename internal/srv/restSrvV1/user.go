package restSrvV1

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jypelle/mifasol/internal/srv/oldstore"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (s *RestServer) readUsers(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Read users")

	users, err := s.oldStore.ReadUsers(nil, &restApiV1.UserFilter{})
	if err != nil {
		logrus.Panicf("Unable to read users: %v", err)
	}

	tool.WriteJsonResponse(w, users)
}

func (s *RestServer) readUser(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Read user")

	vars := mux.Vars(r)
	userId := restApiV1.UserId(vars["id"])

	logrus.Debugf("Read user: %s", userId)

	user, err := s.oldStore.ReadUser(nil, userId)
	if err != nil {
		if err == oldstore.ErrNotFound {
			s.apiErrorCodeResponse(w, restApiV1.NotFoundErrorCode)
			return
		}
		logrus.Panicf("Unable to read user: %v", err)
	}

	tool.WriteJsonResponse(w, user)
}

func (s *RestServer) createUser(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Create user")

	var userMetaComplete restApiV1.UserMetaComplete
	err := json.NewDecoder(r.Body).Decode(&userMetaComplete)
	if err != nil {
		logrus.Panicf("Unable to interpret data to create the user: %v", err)
	}

	user, err := s.oldStore.CreateUser(nil, &userMetaComplete)
	if err != nil {
		logrus.Panicf("Unable to create the user: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	tool.WriteJsonResponse(w, user)
}

func (s *RestServer) updateUser(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	userId := restApiV1.UserId(vars["id"])

	logrus.Debugf("Update user: %s", userId)

	var userMetaComplete restApiV1.UserMetaComplete
	err := json.NewDecoder(r.Body).Decode(&userMetaComplete)
	if err != nil {
		logrus.Panicf("Unable to interpret data to update the user: %v", err)
	}

	user, err := s.oldStore.UpdateUser(nil, userId, &userMetaComplete)
	if err != nil {
		logrus.Panicf("Unable to update the user: %v", err)
	}

	tool.WriteJsonResponse(w, user)

}

func (s *RestServer) deleteUser(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	userId := restApiV1.UserId(vars["id"])

	logrus.Debugf("Delete user: %s", userId)

	user, err := s.oldStore.DeleteUser(nil, userId)
	if err != nil {
		logrus.Panicf("Unable to delete user: %v", err)
	}

	tool.WriteJsonResponse(w, user)

}
