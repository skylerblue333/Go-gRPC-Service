package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const serviceName = "sky-rpc-core"

type ServiceState struct {
	processed atomic.Uint64
	started   time.Time
}

func newState() *ServiceState { return &ServiceState{started: time.Now()} }

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func writeJSON(w http.ResponseWriter, code int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(value)
}

func newHTTPHandler(state *ServiceState) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "service": serviceName})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ready"})
	})
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"processed":      state.processed.Load(),
			"uptime_seconds": int64(time.Since(state.started).Seconds()),
		})
	})
	mux.HandleFunc("POST /v1/process", func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > 1<<20 {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "payload too large"})
			return
		}
		state.processed.Add(1)
		writeJSON(w, http.StatusAccepted, map[string]any{"status": "accepted", "processed": state.processed.Load()})
	})
	return mux
}

func unaryInterceptor(maxConcurrent int64, inFlight *atomic.Int64) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		current := inFlight.Add(1)
		defer inFlight.Add(-1)
		if current > maxConcurrent {
			return nil, status.Error(codes.ResourceExhausted, "server concurrency limit reached")
		}
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("x-request-id"); len(values) > 0 && len(values[0]) > 128 {
				return nil, status.Error(codes.InvalidArgument, "x-request-id too long")
			}
		}
		return handler(ctx, req)
	}
}

func newGRPCServer(maxConcurrent int64) *grpc.Server {
	var inFlight atomic.Int64
	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(1<<20),
		grpc.MaxSendMsgSize(1<<20),
		grpc.ConnectionTimeout(5*time.Second),
		grpc.UnaryInterceptor(unaryInterceptor(maxConcurrent, &inFlight)),
	)
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	reflection.Register(server)
	return server
}

func main() {
	state := newState()
	maxConcurrent, err := strconv.ParseInt(env("MAX_CONCURRENT_RPCS", "256"), 10, 64)
	if err != nil || maxConcurrent < 1 {
		log.Fatal("MAX_CONCURRENT_RPCS must be a positive integer")
	}

	grpcAddr := env("GRPC_ADDR", ":9090")
	httpAddr := env("HTTP_ADDR", ":8080")
	grpcServer := newGRPCServer(maxConcurrent)
	httpServer := &http.Server{
		Addr:              httpAddr,
		Handler:           newHTTPHandler(state),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("gRPC listen: %v", err)
	}

	go func() {
		log.Printf("gRPC listening on %s", grpcAddr)
		if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Printf("gRPC server stopped: %v", err)
		}
	}()
	go func() {
		log.Printf("HTTP operations API listening on %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server stopped: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	done := make(chan struct{})
	go func() { grpcServer.GracefulStop(); close(done) }()
	select {
	case <-done:
	case <-ctx.Done():
		grpcServer.Stop()
	}
	_ = httpServer.Shutdown(ctx)
	_ = grpcListener.Close()
	log.Println("servers stopped cleanly")
}
