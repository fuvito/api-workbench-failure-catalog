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
		id := parseID(r.URL.Path, "/products/")
		p, ok := products[id]
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, p)
	})

	http.ListenAndServe(":19013", mux)
}
