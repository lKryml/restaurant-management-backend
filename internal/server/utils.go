package server

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	jsonResp, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
		return
	}

	_, _ = w.Write(jsonResp)
}

func HandleError(w http.ResponseWriter, status int, message string) {
	WriteJSONResponse(w, status, map[string]string{"error": message})

}
