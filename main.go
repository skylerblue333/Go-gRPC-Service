package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type ServiceState struct {
	mu        sync.RWMutex
	Processed int
	Domain    string
}

var state = &ServiceState{Domain: "skycoin-grpc-service"}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	state.mu.RLock()
	defer state.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"domain":    state.Domain,
		"processed": state.Processed,
	})
}

func handleProcess(w http.ResponseWriter, _ *http.Request) {
	state.mu.Lock()
	state.Processed++
	state.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = fmt.Fprint(w, `{"status":"processing"}`)
}

func newHTTPServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/process", handleProcess)
	return &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func newGRPCServer() *grpc.Server {
	server := grpc.NewServer()
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	reflection.Register(server)
	return server
}

func main() {
	grpcServer := newGRPCServer()
	httpServer := newHTTPServer()

	grpcListener, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("gRPC listen: %v", err)
	}

	go func() {
		log.Println("gRPC server listening on :9090")
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Printf("gRPC server stopped: %v", err)
		}
	}()

	go func() {
		log.Println("HTTP compatibility server listening on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server stopped: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	grpcServer.GracefulStop()
	_ = httpServer.Shutdown(ctx)
	_ = grpcListener.Close()
	log.Println("servers stopped cleanly")
}
