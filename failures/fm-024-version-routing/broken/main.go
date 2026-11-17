package main

import (
	"encoding/json"
	"net/http"
)

// BUG: /v2/products returns the v1 response schema — missing the required
// "sku" field introduced in v2. The router forwards /v2/ to the same
// handler as /v1/ without upgrading the response shape.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type ProductV1 struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	// missing: sku (required in v2)
}

var products = []ProductV1{
	{ID: 1, Name: "Widget", Price: 9.99},
	{ID: 2, Name: "Gadget", Price: 24.99},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// BUG: /v2/products uses the v1 handler — sku field never populated
	mux.HandleFunc("/v2/products", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"items": products})
	})

	http.ListenAndServe(":19024", mux)
}
