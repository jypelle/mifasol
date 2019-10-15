package restSrvV1

import (
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"mifasol/restApiV1"
	"mifasol/tool"
	"net/http"
	"strconv"
)

func (s *RestServer) readSyncReport(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Read sync report")

	vars := mux.Vars(r)
	fromTs, err := strconv.ParseInt(vars["fromTs"], 10, 64)
	if err != nil {
		logrus.Warningf("Unable to interpret timestamp: %v", err)
		s.apiErrorCodeResponse(w, restApiV1.InvalideRequestErrorCode)
		return
	}

	syncReport, err := s.service.ReadSyncReport(fromTs)
	if err != nil {
		logrus.Panicf("Unable to read sync report: %v", err)
	}

	tool.WriteJsonResponse(w, syncReport)
}

func (s *RestServer) readFileSyncReport(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Read file sync report")

	vars := mux.Vars(r)
	fromTs, err := strconv.ParseInt(vars["fromTs"], 10, 64)
	if err != nil {
		logrus.Warningf("Unable to interpret timestamp: %v", err)
		s.apiErrorCodeResponse(w, restApiV1.InvalideRequestErrorCode)
		return
	}

	fileSyncReport, err := s.service.ReadFileSyncReport(fromTs)
	if err != nil {
		logrus.Panicf("Unable to read sync report: %v", err)
	}

	tool.WriteJsonResponse(w, fileSyncReport)
}
