package restSrvV1

import (
	"encoding/json"
	"github.com/jypelle/mifasol/restApiV1"
	"net/http"
)

func (s *RestServer) apiErrorCodeResponse(w http.ResponseWriter, apiErrorCode restApiV1.ErrorCode) {
	apiError := restApiV1.ApiError{ErrorCode: apiErrorCode}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErrorCode.StatusCode())
	json.NewEncoder(w).Encode(apiError)
}

func (s *RestServer) apiErrorResponse(w http.ResponseWriter, apiError restApiV1.ApiError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiError.ErrorCode.StatusCode())
	json.NewEncoder(w).Encode(apiError)
}
