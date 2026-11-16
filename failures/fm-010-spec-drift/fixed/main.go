package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// Field names restored to match the OpenAPI spec.
type Item struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"` // matches spec
	InStock  bool   `json:"in_stock"` // matches spec
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

var items = map[int]Item{
	1: {ID: 1, Name: "Wireless Headphones", Category: "electronics", InStock: true},
	2: {ID: 2, Name: "Desk Chair", Category: "furniture", InStock: false},
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
