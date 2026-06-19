package main

import (
	"context"
	"testing"
)

func TestSayHello(t *testing.T) {
	svc := &GreeterService{}

	resp, err := svc.SayHello(context.Background(), GreetRequest{Name: "World"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Message != "Hello, World!" {
		t.Errorf("unexpected message: %s", resp.Message)
	}
}

func TestSayHelloEmptyName(t *testing.T) {
	svc := &GreeterService{}
	_, err := svc.SayHello(context.Background(), GreetRequest{Name: ""})
	if err == nil {
		t.Error("expected error for empty name")
	}
}
