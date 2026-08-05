package httpapi

import (
	"encoding/json"
	"net/http"
)

type serviceResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", handleServiceInfo)

	return mux
}

func handleServiceInfo(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(serviceResponse{
		Service: "golden-path-service",
		Status:  "running",
	}); err != nil {
		return
	}
}
