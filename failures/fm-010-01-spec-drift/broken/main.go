package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// A developer refactored the domain model field names but did not update the
// OpenAPI spec. The spec says "category" and "in_stock"; the server now emits
// "product_type" and "available" — classic spec/implementation drift.
type Item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ProductType string `json:"product_type"` // spec says: category
	Available   bool   `json:"available"`    // spec says: in_stock
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

var items = map[int]Item{
	1: {ID: 1, Name: "Wireless Headphones", ProductType: "electronics", Available: true},
	2: {ID: 2, Name: "Desk Chair", ProductType: "furniture", Available: false},
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid id", Code: "INVALID_ID"})
			return
		}

		item, ok := items[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not found", Code: "ITEM_NOT_FOUND"})
			return
		}

		writeJSON(w, http.StatusOK, item)
	})

	log.Fatal(http.ListenAndServe(":19010", mux))
}
