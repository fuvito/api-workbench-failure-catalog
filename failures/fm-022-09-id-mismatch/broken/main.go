package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// BUG: handler ignores the path ID and always returns the first product.
// GET /products/5 returns a product with id:1.

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

var products = map[int]Product{
	1: {ID: 1, Name: "Widget", Price: 9.99},
	2: {ID: 2, Name: "Gadget", Price: 19.99},
	3: {ID: 3, Name: "Doohickey", Price: 4.99},
	5: {ID: 5, Name: "Thingamajig", Price: 14.99},
}

func parseID(path, prefix string) int {
	s := strings.TrimPrefix(path, prefix)
	id := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			id = id*10 + int(c-'0')
		}
	}
	return id
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		// BUG: id parsed but never used in lookup; always returns product 1
		_ = parseID(r.URL.Path, "/products/")
		writeJSON(w, http.StatusOK, products[1])
	})

	http.ListenAndServe(":19013", mux)
}
