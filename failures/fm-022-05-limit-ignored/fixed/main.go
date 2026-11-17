package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

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
		result := allItems
		if lStr := r.URL.Query().Get("limit"); lStr != "" {
			if l, err := strconv.Atoi(lStr); err == nil && l >= 0 && l < len(result) {
				result = result[:l]
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"items": result,
			"count": len(result),
		})
	})

	http.ListenAndServe(":19040", mux)
}
