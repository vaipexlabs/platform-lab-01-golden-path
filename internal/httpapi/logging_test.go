package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStructuredRequestLog(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	handler := NewHandler(logger)

	handler.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
	)

	var event map[string]any
	if err := json.Unmarshal(output.Bytes(), &event); err != nil {
		t.Fatalf("decode log event: %v", err)
	}

	expected := map[string]any{
		"level":  "INFO",
		"msg":    "http request completed",
		"method": http.MethodGet,
		"route":  "/",
		"status": float64(http.StatusOK),
	}

	for field, want := range expected {
		if got := event[field]; got != want {
			t.Errorf("expected %s to be %v, got %v", field, want, got)
		}
	}

	if _, exists := event["duration_ms"]; !exists {
		t.Error("expected duration_ms field")
	}
}

func TestStructuredRequestLogUsesBoundedRoute(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	handler := NewHandler(logger)

	handler.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/customers/12345", nil),
	)

	if strings.Contains(output.String(), "/customers/12345") {
		t.Error("expected request log to avoid the raw URL")
	}

	if !strings.Contains(output.String(), `"route":"unmatched"`) {
		t.Error("expected unmatched route label")
	}
}

func TestSuccessfulOperationalRequestsAreNotLogged(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	handler := NewHandler(logger)

	for _, path := range []string{"/health/live", "/health/ready", "/metrics"} {
		handler.ServeHTTP(
			httptest.NewRecorder(),
			httptest.NewRequest(http.MethodGet, path, nil),
		)
	}

	if output.Len() != 0 {
		t.Fatalf("expected no successful operational request logs, got %s", output.String())
	}
}
