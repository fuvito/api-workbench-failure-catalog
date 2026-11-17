package main

import (
	"encoding/json"
	"net/http"
)

// BUG: the limit query parameter is parsed but never applied.
// GET /items?limit=2 returns all 5 items instead of 2.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var allItems = []Item{
	{ID: 1, Name: "Alpha"},
	{ID: 2, Name: "Beta"},
	{ID: 3, Name: "Gamma"},
	{ID: 4, Name: "Delta"},
	{ID: 5, Name: "Epsilon"},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		// BUG: limit is parsed but never used to slice allItems
		_ = r.URL.Query().Get("limit")
		writeJSON(w, http.StatusOK, map[string]any{
			"items": allItems,
			"count": len(allItems),
		})
	})

	http.ListenAndServe(":19040", mux)
}
