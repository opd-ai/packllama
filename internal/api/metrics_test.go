package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsEndpoint_Disabled(t *testing.T) {
	handler := newHandlerWithConfig(Config{Logger: testLogger(io.Discard)}, nil)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	// /metrics should not be routed when EnableMetrics=false.
	// net/http returns 404 for unregistered patterns.
	if recorder.Code == http.StatusOK {
		t.Fatal("expected /metrics to be unavailable when EnableMetrics=false")
	}
}

func TestMetricsEndpoint_Enabled(t *testing.T) {
	handler := newHandlerWithConfig(Config{
		Logger:        testLogger(io.Discard),
		EnableMetrics: true,
	}, nil)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 for /metrics, got %d", recorder.Code)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "packllama_http_request_duration_seconds") {
		t.Fatalf("expected packllama metrics in response, got:\n%s", body)
	}
}

func TestMetricsCountsRequests(t *testing.T) {
	handler := newHandlerWithConfig(Config{
		Logger:        testLogger(io.Discard),
		EnableMetrics: true,
	}, nil)

	// Make a request to /health to increment the counter.
	healthRecorder := httptest.NewRecorder()
	handler.ServeHTTP(healthRecorder, httptest.NewRequest(http.MethodGet, "/health", nil))

	// Check the metrics endpoint now.
	metricsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(metricsRecorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	body := metricsRecorder.Body.String()
	if !strings.Contains(body, `method="GET"`) {
		t.Fatalf("expected method label in metrics, got:\n%s", body)
	}
}
