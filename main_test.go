package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	handleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", resp["status"])
	}
	if resp["domain"] != "skycoin-grpc-service" {
		t.Errorf("unexpected domain: %v", resp["domain"])
	}
}

func TestProcess(t *testing.T) {
	state.mu.Lock()
	initial := state.Processed
	state.mu.Unlock()

	req, _ := http.NewRequest("POST", "/process", nil)
	rr := httptest.NewRecorder()
	handleProcess(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("expected 202 Accepted, got %d", rr.Code)
	}

	state.mu.RLock()
	processed := state.Processed
	state.mu.RUnlock()
	if processed != initial+1 {
		t.Errorf("expected state to increment from %d to %d", initial, initial+1)
	}
}

func TestGRPCServer(t *testing.T) {
	server := newGRPCServer()
	if server == nil {
		t.Fatal("expected gRPC server")
	}
	server.Stop()
}
