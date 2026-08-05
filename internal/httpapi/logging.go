package httpapi

import (
	"log/slog"
	"net/http"
	"time"
)

func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		response := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(response, r)

		route := requestRoute(r)
		if suppressSuccessfulOperationalLog(route, response.statusCode) {
			return
		}

		logger.InfoContext(
			r.Context(),
			"http request completed",
			"method", r.Method,
			"route", route,
			"status", response.statusCode,
			"duration_ms", float64(time.Since(started).Microseconds())/1000,
		)
	})
}

func suppressSuccessfulOperationalLog(route string, statusCode int) bool {
	if statusCode >= http.StatusBadRequest {
		return false
	}

	switch route {
	case "/health/live", "/health/ready", "/metrics":
		return true
	default:
		return false
	}
}
