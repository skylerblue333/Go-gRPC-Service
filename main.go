package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
)

// Simplified gRPC-style service without proto dependency for portability
type GreetRequest struct {
	Name string
}

type GreetResponse struct {
	Message string
}

type GreeterService struct{}

func (g *GreeterService) SayHello(ctx context.Context, req GreetRequest) (GreetResponse, error) {
	if req.Name == "" {
		return GreetResponse{}, fmt.Errorf("name cannot be empty")
	}
	return GreetResponse{Message: "Hello, " + req.Name + "!"}, nil
}

func main() {
	svc := &GreeterService{}
	mux := http.NewServeMux()

	mux.HandleFunc("/greet", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		resp, err := svc.SayHello(r.Context(), GreetRequest{Name: name})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprint(w, resp.Message)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	ln, _ := net.Listen("tcp", ":8080")
	log.Println("gRPC-style service on :8080")
	log.Fatal(http.Serve(ln, mux))
}
