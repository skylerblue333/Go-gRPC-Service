package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthAndReadiness(t *testing.T) {
	h := newHTTPHandler(newState())
	for _, path := range []string{"/healthz", "/readyz"} {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", path, rr.Code)
		}
	}
}

func TestProcessAndMetrics(t *testing.T) {
	state := newState()
	h := newHTTPHandler(state)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/v1/process", strings.NewReader("{}")))
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}
	if state.processed.Load() != 1 {
		t.Fatalf("expected processed count 1")
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	var response map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response["processed"].(float64) != 1 {
		t.Fatalf("unexpected metric: %v", response["processed"])
	}
}

func TestPayloadLimit(t *testing.T) {
	h := newHTTPHandler(newState())
	req := httptest.NewRequest(http.MethodPost, "/v1/process", nil)
	req.ContentLength = (1 << 20) + 1
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rr.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	h := newHTTPHandler(newState())
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/process", nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestGRPCServer(t *testing.T) {
	server := newGRPCServer(16)
	if server == nil {
		t.Fatal("expected gRPC server")
	}
	server.Stop()
}
