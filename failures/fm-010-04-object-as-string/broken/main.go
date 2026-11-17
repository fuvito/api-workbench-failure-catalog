package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// BUG: metadata is serialised as a JSON-encoded string rather than a nested object.
// The struct field type is string; Go marshals it as a JSON string literal.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type DocumentBroken struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Metadata string `json:"metadata"` // BUG: should be map[string]any
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/documents/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/documents/") {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, DocumentBroken{
			ID:       1,
			Title:    "API Design Guide",
			Metadata: `{"author":"Alice","version":1}`,
		})
	})

	http.ListenAndServe(":19021", mux)
}
