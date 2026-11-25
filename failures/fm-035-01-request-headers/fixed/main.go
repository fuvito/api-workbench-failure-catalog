package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

// Request headers are parsed as HTTP defines them, validated, and either
// honoured or rejected with a documented status.

type document struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// Synthetic bearer token.
const demoToken = "tok-alice"

var correlationID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

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

// matches reports whether an If-Match or If-None-Match value lists tag or is
// "*". With weak set, W/"x" matches "x" (If-None-Match); otherwise a weak tag
// never matches (If-Match).
func matches(header, tag string, weak bool) bool {
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimSpace(candidate)
		if weak {
			candidate = strings.TrimPrefix(candidate, "W/")
		}
		if candidate == "*" || candidate == tag {
			return true
		}
	}
	return false
}

func withHeaders(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Correlation-ID")
		valid := correlationID.MatchString(id)
		if !valid {
			// A new ID: an invalid one is never echoed.
			w.Header().Set("X-Correlation-ID", rand.Text())
		} else {
			w.Header().Set("X-Correlation-ID", id)
		}
		if id != "" && !valid {
			writeError(w, http.StatusBadRequest, "invalid X-Correlation-ID")
			return
		}

		// The scheme is case-insensitive; the token is not.
		scheme, token, _ := strings.Cut(r.Header.Get("Authorization"), " ")
		if !strings.EqualFold(scheme, "Bearer") || token != demoToken {
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
	if matches(r.Header.Get("If-None-Match"), etag(), true) {
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
	ifMatch := r.Header.Get("If-Match")
	if ifMatch == "" {
		writeError(w, http.StatusPreconditionRequired, "If-Match is required")
		return
	}
	// Check and write under one lock, so the version can't change in between.
	mu.Lock()
	defer mu.Unlock()

	if !matches(ifMatch, etag(), false) {
		writeError(w, http.StatusPreconditionFailed, "the document has changed")
		return
	}
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
