package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServiceInfo(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	NewHandler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", contentType)
	}

	var got serviceResponse
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	want := serviceResponse{
		Service: "golden-path-service",
		Status:  "running",
	}

	if got != want {
		t.Fatalf("expected response %+v, got %+v", want, got)
	}
}

func TestHealthEndpoints(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		status string
	}{
		{name: "liveness", path: "/health/live", status: "alive"},
		{name: "readiness", path: "/health/ready", status: "ready"},
	}

	handler := NewHandler()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
			}

			if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
				t.Fatalf("expected Content-Type application/json, got %q", contentType)
			}

			var got healthResponse
			if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			want := healthResponse{Status: test.status}
			if got != want {
				t.Fatalf("expected response %+v, got %+v", want, got)
			}
		})
	}
}

func TestMetrics(t *testing.T) {
	handler := NewHandler()

	requests := []*http.Request{
		httptest.NewRequest(http.MethodGet, "/", nil),
		httptest.NewRequest(http.MethodGet, "/customers/12345", nil),
	}

	for _, request := range requests {
		handler.ServeHTTP(httptest.NewRecorder(), request)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/plain") {
		t.Fatalf("expected Prometheus text content, got %q", contentType)
	}

	body := response.Body.String()
	expectedMetrics := []string{
		"go_goroutines",
		"process_cpu_seconds_total",
		`golden_path_http_requests_total{method="GET",route="/",status="200"} 1`,
		`golden_path_http_requests_total{method="GET",route="unmatched",status="404"} 1`,
		`golden_path_http_request_duration_seconds_count{method="GET",route="/"} 1`,
		"golden_path_http_requests_in_flight 1",
	}

	for _, metric := range expectedMetrics {
		if !strings.Contains(body, metric) {
			t.Errorf("expected response to contain %q", metric)
		}
	}

	if strings.Contains(body, "/customers/12345") {
		t.Error("expected unmatched requests to avoid raw-path metric labels")
	}
}
