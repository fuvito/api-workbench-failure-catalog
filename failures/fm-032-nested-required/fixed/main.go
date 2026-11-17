package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code,omitempty"`
}

type Order struct {
	ID              int     `json:"id"`
	Total           float64 `json:"total"`
	ShippingAddress Address `json:"shipping_address"`
}

var orders = map[int]Order{
	1: {ID: 1, Total: 49.99, ShippingAddress: Address{Street: "123 Main St", City: "New York", Country: "US"}},
	2: {ID: 2, Total: 29.99, ShippingAddress: Address{Street: "456 Elm Ave", City: "Boston", Country: "US"}},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/orders/"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}
		o, ok := orders[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusOK, o)
	})

	http.ListenAndServe(":19032", mux)
}
