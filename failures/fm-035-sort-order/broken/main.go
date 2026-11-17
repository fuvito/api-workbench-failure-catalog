package main

import (
	"encoding/json"
	"net/http"
)

// BUG: the sort query parameter is accepted but completely ignored.
// Products are always returned in insertion order regardless of ?sort=.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

var products = []Product{
	{ID: 2, Name: "Keyboard", Price: 79.99},
	{ID: 3, Name: "Monitor", Price: 349.99},
	{ID: 1, Name: "Mouse", Price: 29.99},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		// BUG: sort parameter read but ignored — always returns insertion order
		_ = r.URL.Query().Get("sort")
		writeJSON(w, http.StatusOK, map[string]any{"items": products})
	})

	http.ListenAndServe(":19035", mux)
}
