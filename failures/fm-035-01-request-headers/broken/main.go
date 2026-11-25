package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// BUG: request headers are mishandled. The auth scheme is matched
// case-sensitively, the correlation ID is lowercased and never validated,
// If-None-Match must be one exact value, and If-Match is ignored.

type document struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// Synthetic bearer token.
const demoToken = "tok-alice"

var (
	mu      sync.Mutex
	version = 1
	record  = document{ID: "doc-1001", Title: "First draft"}
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// etag returns the current entity tag. The caller must hold mu.
func etag() string { return fmt.Sprintf(`"v%d"`, version) }

func withHeaders(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// BUG: changes the caller's ID, and accepts any value
		if id := r.Header.Get("X-Correlation-ID"); id != "" {
			w.Header().Set("X-Correlation-ID", strings.ToLower(id))
		}
		// BUG: "bearer" and "BEARER" are the same scheme as "Bearer"
		if r.Header.Get("Authorization") != "Bearer "+demoToken {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r)
	}
}

func getDocument(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("ETag", etag())
	// BUG: the header may hold a list, and entries may be weak (W/"v1")
	if r.Header.Get("If-None-Match") == etag() {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func updateDocument(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	mu.Lock()
	defer mu.Unlock()

	// BUG: If-Match is never read
	record.Title = body.Title
	version++
	w.Header().Set("ETag", etag())
	writeJSON(w, http.StatusOK, record)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /documents/doc-1001", withHeaders(getDocument))
	mux.HandleFunc("PUT /documents/doc-1001", withHeaders(updateDocument))

	http.ListenAndServe(":19059", mux)
}
