package main

import (
	"encoding/json"
	"net/http"
	"sort"
)

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
		result := make([]Product, len(products))
		copy(result, products)

		switch r.URL.Query().Get("sort") {
		case "price_asc":
			sort.Slice(result, func(i, j int) bool { return result[i].Price < result[j].Price })
		case "price_desc":
			sort.Slice(result, func(i, j int) bool { return result[i].Price > result[j].Price })
		case "name_asc":
			sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
		}

		writeJSON(w, http.StatusOK, map[string]any{"items": result})
	})

	http.ListenAndServe(":19035", mux)
}
