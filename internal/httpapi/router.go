package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type serviceResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

type healthResponse struct {
	Status string `json:"status"`
}

func NewHandler(logger *slog.Logger) http.Handler {
	metrics := newHTTPMetrics()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", handleServiceInfo)
	mux.HandleFunc("GET /health/live", handleLiveness)
	mux.HandleFunc("GET /health/ready", handleReadiness)
	mux.Handle("GET /metrics", metrics.handler())

	return logRequests(logger, metrics.instrument(mux))
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

func handleLiveness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(healthResponse{Status: "alive"}); err != nil {
		return
	}
}

func handleReadiness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(healthResponse{Status: "ready"}); err != nil {
		return
	}
}
