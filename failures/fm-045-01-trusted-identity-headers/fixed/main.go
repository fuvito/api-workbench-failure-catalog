package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// The backend verifies the bearer token itself and takes identity only from it.
// The gateway rejects requests without a valid token and removes identity
// headers sent by the caller.
//
// One process runs both services: the gateway on :19056 and the backend on
// :19057. The backend port stands for an internal service a caller can reach.

type user struct {
	ID   string `json:"user_id"`
	Role string `json:"role"`
}

// Synthetic bearer tokens.
var userByToken = map[string]user{
	"tok-alice": {ID: "alice", Role: "member"},
	"tok-carol": {ID: "carol", Role: "admin"},
}

// Headers that must never be accepted from a caller.
var identityHeaders = []string{"X-User-ID", "X-User-Role"}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func verifyToken(r *http.Request) (user, bool) {
	token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	u, known := userByToken[token]
	return u, found && known
}

// ---- backend ----

func identify(w http.ResponseWriter, r *http.Request) (user, bool) {
	u, ok := verifyToken(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return user{}, false
	}
	return u, true
}

func backend() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		if u, ok := identify(w, r); ok {
			writeJSON(w, http.StatusOK, u)
		}
	})
	mux.HandleFunc("GET /admin/users", func(w http.ResponseWriter, r *http.Request) {
		u, ok := identify(w, r)
		if !ok {
			return
		}
		if u.Role != "admin" {
			writeError(w, http.StatusForbidden, "admin role required")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"users": []string{"alice", "carol"}})
	})
	return mux
}

// ---- gateway ----

func gateway(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, h := range identityHeaders {
			r.Header.Del(h)
		}
		if _, ok := verifyToken(r); !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Listen on the backend port first, so it is ready when /health answers.
	backendListener, err := net.Listen("tcp", ":19057")
	if err != nil {
		log.Fatal(err)
	}
	go func() { log.Fatal(http.Serve(backendListener, backend())) }()

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: "localhost:19057"})
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/", gateway(proxy))

	log.Fatal(http.ListenAndServe(":19056", mux))
}
