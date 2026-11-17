package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price string `json:"price"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

var products = map[int]Product{
	1: {ID: 1, Name: "Widget", Price: "9.99"},
	2: {ID: 2, Name: "Gadget", Price: "24.99"},
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

	mux.HandleFunc("GET /products/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid id", Code: "INVALID_ID"})
			return
		}

		product, ok := products[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, ErrorResponse{
				Error: "not found",
				Code:  "PRODUCT_NOT_FOUND",
			})
			return
		}

		writeJSON(w, http.StatusOK, product)
	})

	log.Fatal(http.ListenAndServe(":19025", mux))
}
