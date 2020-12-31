package tool

import (
	"encoding/json"
	"net/http"
)

func WriteJsonResponse(w http.ResponseWriter, object interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(object)
}
