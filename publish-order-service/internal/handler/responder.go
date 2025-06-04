package handler

import (
	"encoding/json"
	"net/http"
)

type response map[string]interface{}

func writeJSON(w http.ResponseWriter, status int, body response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
