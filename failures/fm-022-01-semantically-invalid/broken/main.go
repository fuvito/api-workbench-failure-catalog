package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// BUG: total is set to unit_price instead of unit_price * quantity.
// The response is syntactically valid JSON with all required fields at the
// correct types — schema validation passes. The semantic error (wrong math)
// is only caught by asserting the computed value.
type Order struct {
	ID        int     `json:"id"`
	Customer  string  `json:"customer"`
	Item      string  `json:"item"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Total     float64 `json:"total"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

var orders = map[int]Order{
	1: {ID: 1, Customer: "Alice Corp", Item: "Widget", Quantity: 3, UnitPrice: 10.00,
		Total: 10.00},  // BUG: should be 30.00 (3 × 10.00)
	2: {ID: 2, Customer: "Bob Ltd", Item: "Gadget", Quantity: 2, UnitPrice: 24.99,
		Total: 24.99},  // BUG: should be 49.98 (2 × 24.99)
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

	mux.HandleFunc("GET /orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid id", Code: "INVALID_ID"})
			return
		}

		order, ok := orders[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not found", Code: "ORDER_NOT_FOUND"})
			return
		}

		writeJSON(w, http.StatusOK, order)
	})

	log.Fatal(http.ListenAndServe(":19022", mux))
}
