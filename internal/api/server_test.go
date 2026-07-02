package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/opd-ai/packllama/internal/service"
)

func TestHealthEndpoint(t *testing.T) {
	handler := NewHandler(testLogger(io.Discard), []string{"http://localhost:3000"}, nil)
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	request.Header.Set("Origin", "http://localhost:3000")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if recorder.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("expected CORS header to be echoed back")
	}
	if recorder.Header().Get("X-Request-ID") == "" {
		t.Fatalf("expected request ID header to be set")
	}

	var payload service.HealthStatus
	if err := json.NewDecoder(recorder.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected health status ok, got %q", payload.Status)
	}
}

func TestLoggingIncludesRequestMetadata(t *testing.T) {
	var logOutput bytes.Buffer
	logger := testLogger(&logOutput)
	handler := NewHandler(logger, nil, nil)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))

	logLine := logOutput.String()
	if !strings.Contains(logLine, "\"request_id\"") {
		t.Fatalf("expected request_id in log output: %s", logLine)
	}
	if !strings.Contains(logLine, "\"status\":200") {
		t.Fatalf("expected status in log output: %s", logLine)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	handler := chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}),
		withRequestID,
		withLogging(testLogger(io.Discard)),
		withRecovery(testLogger(io.Discard)),
	)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "internal server error") {
		t.Fatalf("expected error payload, got %q", recorder.Body.String())
	}
}

func TestServerStartAndShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := NewServer(Config{
		Host:            "127.0.0.1",
		Port:            0,
		ShutdownTimeout: time.Second,
		Logger:          testLogger(io.Discard),
	}, nil)

	errs := make(chan error, 1)
	go func() {
		errs <- server.Start(ctx)
	}()

	addr := waitForAddr(t, server)
	response, err := http.Get("http://" + addr + "/health")
	if err != nil {
		t.Fatalf("request health endpoint: %v", err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}

	cancel()

	select {
	case err := <-errs:
		if err != nil {
			t.Fatalf("server exited with error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for server shutdown")
	}
}

func waitForAddr(t *testing.T, server *Server) string {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		addr := server.ListenAddr()
		if !strings.HasSuffix(addr, ":0") {
			return addr
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server did not start listening in time")
	return ""
}

func testLogger(writer io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo}))
}
