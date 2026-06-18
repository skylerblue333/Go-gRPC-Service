package main

import (
	"context"
	"log"
	"net"
	"google.golang.org/grpc"
)

// UserService implements a simple user lookup service
type UserService struct{}

// GetUser simulates a database lookup
func (s *UserService) GetUser(ctx context.Context, req *UserRequest) (*UserResponse, error) {
	log.Printf("GetUser called for ID: %s", req.UserId)
	return &UserResponse{
		UserId: req.UserId,
		Name:   "Skyler Blue",
		Email:  "skyler@example.com",
	}, nil
}

// UserRequest is the request message
type UserRequest struct {
	UserId string
}

// UserResponse is the response message
type UserResponse struct {
	UserId string
	Name   string
	Email  string
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	
	s := grpc.NewServer()
	log.Println("gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
