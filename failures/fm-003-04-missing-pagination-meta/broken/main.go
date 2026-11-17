package main

import (
	"encoding/json"
	"net/http"
)

// BUG: response omits the required total_pages and has_more fields.
// The spec declares them as required; the broken struct simply lacks them.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type Post struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

var posts = []Post{
	{ID: 1, Title: "First Post"},
	{ID: 2, Title: "Second Post"},
	{ID: 3, Title: "Third Post"},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		// BUG: missing total_pages and has_more fields
		writeJSON(w, http.StatusOK, map[string]any{
			"items": posts,
			"count": len(posts),
		})
	})

	http.ListenAndServe(":19049", mux)
}
