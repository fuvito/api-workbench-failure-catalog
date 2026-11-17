package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type DocumentFixed struct {
	ID       int            `json:"id"`
	Title    string         `json:"title"`
	Metadata map[string]any `json:"metadata"`
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
		writeJSON(w, http.StatusOK, DocumentFixed{
			ID:    1,
			Title: "API Design Guide",
			Metadata: map[string]any{
				"author":  "Alice",
				"version": 1,
			},
		})
	})

	http.ListenAndServe(":19021", mux)
}
