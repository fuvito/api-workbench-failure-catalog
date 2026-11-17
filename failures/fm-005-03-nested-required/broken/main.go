package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// BUG: shipping_address.street is omitted when the address was entered
// without a street line — the field is left as an empty string but the
// spec requires it to be present and non-null.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type AddressBroken struct {
	City       string `json:"city"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code,omitempty"`
	// BUG: street omitted entirely when empty (omitempty drops required field)
	Street string `json:"street,omitempty"`
}

type Order struct {
	ID              int           `json:"id"`
	Total           float64       `json:"total"`
	ShippingAddress AddressBroken `json:"shipping_address"`
}

var orders = map[int]Order{
	1: {ID: 1, Total: 49.99, ShippingAddress: AddressBroken{Street: "123 Main St", City: "New York", Country: "US"}},
	2: {ID: 2, Total: 29.99, ShippingAddress: AddressBroken{Street: "", City: "Boston", Country: "US"}}, // BUG: street omitted
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
