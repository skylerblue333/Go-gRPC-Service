package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestHealthAndReadiness(t *testing.T) {
	h := newHTTPHandler(newState(), newTokenAuthorizer("secret"))
	for _, path := range []string{"/healthz", "/readyz"} {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", path, rr.Code)
		}
	}
}

func TestHTTPAuth(t *testing.T) {
	state := newState()
	h := newHTTPHandler(state, newTokenAuthorizer("secret"))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/v1/process", strings.NewReader("{}")))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", rr.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/process", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer secret")
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202 with token, got %d", rr.Code)
	}
}

func TestProcessAndMetrics(t *testing.T) {
	state := newState()
	h := newHTTPHandler(state, newTokenAuthorizer(""))
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
	h := newHTTPHandler(newState(), newTokenAuthorizer(""))
	req := httptest.NewRequest(http.MethodPost, "/v1/process", nil)
	req.ContentLength = (1 << 20) + 1
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rr.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	h := newHTTPHandler(newState(), newTokenAuthorizer(""))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/process", nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestGRPCServer(t *testing.T) {
	server := newGRPCServer(16, newTokenAuthorizer(""), newState())
	if server == nil {
		t.Fatal("expected gRPC server")
	}
	server.Stop()
}

func TestUnaryInterceptorAuthAndRequestID(t *testing.T) {
	var inFlight atomic.Int64
	state := newState()
	interceptor := unaryInterceptor(1, &inFlight, newTokenAuthorizer("secret"), state)
	handler := func(ctx context.Context, req any) (any, error) { return "ok", nil }
	info := &grpc.UnaryServerInfo{FullMethod: "/sky.v1.Work/Run"}

	_, err := interceptor(context.Background(), nil, info, handler)
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", err)
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer secret", "x-request-id", "req-1"))
	resp, err := interceptor(ctx, nil, info, handler)
	if err != nil || resp != "ok" {
		t.Fatalf("expected authorized success, got resp=%v err=%v", resp, err)
	}

	ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer secret", "x-request-id", strings.Repeat("x", 129)))
	_, err = interceptor(ctx, nil, info, handler)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected invalid request id, got %v", err)
	}
}
