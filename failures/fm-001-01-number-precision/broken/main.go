package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// BUG: monetary fields typed as string — JSON output is "24.00" (string)
// instead of 24.00 (number), violating the declared OpenAPI spec.
type Invoice struct {
	ID       int    `json:"id"`
	Customer string `json:"customer"`
	Subtotal string `json:"subtotal"` // should be float64
	Tax      string `json:"tax"`      // should be float64
	Total    string `json:"total"`    // should be float64
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

var invoices = map[int]Invoice{
	1: {ID: 1, Customer: "Acme Corp", Subtotal: "20.00", Tax: "4.00", Total: "24.00"},
	2: {ID: 2, Customer: "Globex Inc", Subtotal: "49.99", Tax: "10.00", Total: "59.99"},
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

	mux.HandleFunc("GET /invoices/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid id", Code: "INVALID_ID"})
			return
		}

		invoice, ok := invoices[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not found", Code: "INVOICE_NOT_FOUND"})
			return
		}

		writeJSON(w, http.StatusOK, invoice)
	})

	log.Fatal(http.ListenAndServe(":19001", mux))
}
