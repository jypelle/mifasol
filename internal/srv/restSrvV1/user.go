package restSrvV1

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jypelle/mifasol/internal/srv/storeerror"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"net/http"
)

func (s *RestServer) readUsers(w http.ResponseWriter, r *http.Request) {
	s.log.Debugf("Read users")

	users, err := s.store.ReadUsers(nil, &restApiV1.UserFilter{})
	if err != nil {
		s.log.Panicf("Unable to read users: %v", err)
	}

	tool.WriteJsonResponse(w, users)
}

func (s *RestServer) readUser(w http.ResponseWriter, r *http.Request) {
	s.log.Debugf("Read user")

	vars := mux.Vars(r)
	userId := restApiV1.UserId(vars["id"])

	s.log.Debugf("Read user: %s", userId)

	user, err := s.store.ReadUser(nil, userId)
	if err != nil {
		if err == storeerror.ErrNotFound {
			s.apiErrorCodeResponse(w, restApiV1.NotFoundErrorCode)
			return
		}
		s.log.Panicf("Unable to read user: %v", err)
	}

	tool.WriteJsonResponse(w, user)
}

func (s *RestServer) createUser(w http.ResponseWriter, r *http.Request) {
	s.log.Debugf("Create user")

	var userMetaComplete restApiV1.UserMetaComplete
	err := json.NewDecoder(r.Body).Decode(&userMetaComplete)
	if err != nil {
		s.log.Panicf("Unable to interpret data to create the user: %v", err)
	}

	user, err := s.store.CreateUser(nil, &userMetaComplete)
	if err != nil {
		s.log.Panicf("Unable to create the user: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	tool.WriteJsonResponse(w, user)
}

func (s *RestServer) updateUser(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	userId := restApiV1.UserId(vars["id"])

	s.log.Debugf("Update user: %s", userId)

	var userMetaComplete restApiV1.UserMetaComplete
	err := json.NewDecoder(r.Body).Decode(&userMetaComplete)
	if err != nil {
		s.log.Panicf("Unable to interpret data to update the user: %v", err)
	}

	user, err := s.store.UpdateUser(nil, userId, &userMetaComplete)
	if err != nil {
		s.log.Panicf("Unable to update the user: %v", err)
	}

	tool.WriteJsonResponse(w, user)

}

func (s *RestServer) deleteUser(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	userId := restApiV1.UserId(vars["id"])

	s.log.Debugf("Delete user: %s", userId)

	user, err := s.store.DeleteUser(nil, userId)
	if err != nil {
		s.log.Panicf("Unable to delete user: %v", err)
	}

	tool.WriteJsonResponse(w, user)

}
