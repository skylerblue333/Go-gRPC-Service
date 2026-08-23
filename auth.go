package main

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const authHeader = "authorization"

type tokenAuthorizer struct {
	token string
}

func newTokenAuthorizer(token string) tokenAuthorizer {
	return tokenAuthorizer{token: strings.TrimSpace(token)}
}

func (a tokenAuthorizer) enabled() bool {
	return a.token != ""
}

func (a tokenAuthorizer) matches(value string) bool {
	if !a.enabled() {
		return true
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(value, prefix) {
		return false
	}
	provided := strings.TrimSpace(strings.TrimPrefix(value, prefix))
	if len(provided) != len(a.token) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(a.token)) == 1
}

func (a tokenAuthorizer) authorizeContext(ctx context.Context) error {
	if !a.enabled() {
		return nil
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing authorization metadata")
	}
	values := md.Get(authHeader)
	if len(values) != 1 || !a.matches(values[0]) {
		return status.Error(codes.Unauthenticated, "invalid bearer token")
	}
	return nil
}

func (a tokenAuthorizer) httpMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			next.ServeHTTP(w, r)
			return
		}
		if a.enabled() && !a.matches(r.Header.Get("Authorization")) {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
