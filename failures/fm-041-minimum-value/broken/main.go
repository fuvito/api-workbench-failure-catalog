package main

import (
	"encoding/json"
	"net/http"
)

// BUG: products are stored and returned with negative price and stock values,
// violating the minimum: 0 constraint declared in the OpenAPI schema.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

var products = []Product{
	{ID: 1, Name: "Widget", Price: 9.99, Stock: 10},
	{ID: 2, Name: "Gadget", Price: -5.00, Stock: -3}, // BUG: negative values
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"items": products,
			"count": len(products),
		})
	})

	http.ListenAndServe(":19041", mux)
}
