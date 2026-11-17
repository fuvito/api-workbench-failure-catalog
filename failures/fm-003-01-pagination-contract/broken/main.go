package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// BUG: has_more is always true, even on the last page.
// A client that drives pagination by checking has_more will loop forever:
// after the last page it receives an empty items slice with has_more: true
// and re-requests indefinitely.

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type ProductPage struct {
	Items   []Product `json:"items"`
	Total   int       `json:"total"`
	Page    int       `json:"page"`
	Limit   int       `json:"limit"`
	HasMore bool      `json:"has_more"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

var products = []Product{
	{ID: 1, Name: "Widget", Price: 9.99},
	{ID: 2, Name: "Gadget", Price: 24.99},
	{ID: 3, Name: "Doohickey", Price: 4.99},
	{ID: 4, Name: "Thingamajig", Price: 14.99},
	{ID: 5, Name: "Whatsit", Price: 7.99},
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func queryInt(r *http.Request, key string, def int) (int, error) {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def, nil
	}
	return strconv.Atoi(s)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /products", func(w http.ResponseWriter, r *http.Request) {
		page, err := queryInt(r, "page", 1)
		if err != nil || page < 1 {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid page", Code: "INVALID_PAGE"})
			return
		}
		limit, err := queryInt(r, "limit", 10)
		if err != nil || limit < 1 {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid limit", Code: "INVALID_LIMIT"})
			return
		}

		start := (page - 1) * limit
		if start > len(products) {
			start = len(products)
		}
		end := start + limit
		if end > len(products) {
			end = len(products)
		}

		writeJSON(w, http.StatusOK, ProductPage{
			Items:   products[start:end],
			Total:   len(products),
			Page:    page,
			Limit:   limit,
			HasMore: true, // BUG: should be end < len(products)
		})
	})

	log.Fatal(http.ListenAndServe(":19003", mux))
}
