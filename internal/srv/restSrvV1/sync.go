package restSrvV1

import (
	"github.com/gorilla/mux"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"net/http"
	"strconv"
)

func (s *RestServer) readSyncReport(w http.ResponseWriter, r *http.Request) {
	s.log.Debugf("Read sync report")

	vars := mux.Vars(r)
	fromTs, err := strconv.ParseInt(vars["fromTs"], 10, 64)
	if err != nil {
		s.log.Warningf("Unable to interpret timestamp: %v", err)
		s.apiErrorCodeResponse(w, restApiV1.InvalideRequestErrorCode)
		return
	}

	syncReport, err := s.store.ReadSyncReport(fromTs)
	if err != nil {
		s.log.Panicf("Unable to read sync report: %v", err)
	}

	tool.WriteJsonResponse(w, syncReport)
}

func (s *RestServer) readFileSyncReport(w http.ResponseWriter, r *http.Request) {
	s.log.Debugf("Read file sync report")

	vars := mux.Vars(r)
	fromTs, err := strconv.ParseInt(vars["fromTs"], 10, 64)
	if err != nil {
		s.log.Warningf("Unable to interpret timestamp: %v", err)
		s.apiErrorCodeResponse(w, restApiV1.InvalideRequestErrorCode)
		return
	}
	userId := restApiV1.UserId(vars["userId"])

	fileSyncReport, err := s.store.ReadFileSyncReport(fromTs, userId)
	if err != nil {
		s.log.Panicf("Unable to read sync report: %v", err)
	}

	tool.WriteJsonResponse(w, fileSyncReport)
}
