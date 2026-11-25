package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

// BUG: the invoice header and each line are saved one at a time. When a later
// write fails, the request returns an error but the earlier writes remain.

type line struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

type invoice struct {
	Number string `json:"number"`
	Lines  []line `json:"lines"`
}

var (
	mu       sync.Mutex
	invoices = map[string]*invoice{}
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func createInvoice(w http.ResponseWriter, r *http.Request) {
	var in invoice
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Number == "" || len(in.Lines) == 0 {
		writeError(w, http.StatusBadRequest, "number and lines are required")
		return
	}
	// Test hook: the database "fails" while writing the second line.
	failSecondLine := r.Header.Get("X-Test-Fault") == "second-line-write"

	mu.Lock()
	defer mu.Unlock()
	if _, exists := invoices[in.Number]; exists {
		writeError(w, http.StatusConflict, "invoice number already exists")
		return
	}

	// BUG: the header is committed before the lines
	stored := &invoice{Number: in.Number, Lines: []line{}}
	invoices[in.Number] = stored

	seen := map[string]bool{}
	for i, l := range in.Lines {
		if i == 1 && failSecondLine {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		// A unique constraint: one line per SKU.
		if seen[l.SKU] {
			writeError(w, http.StatusConflict, "duplicate sku")
			return
		}
		seen[l.SKU] = true
		stored.Lines = append(stored.Lines, l)
	}
	writeJSON(w, http.StatusCreated, stored)
}

func getInvoice(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	stored, ok := invoices[r.PathValue("number")]
	if !ok {
		writeError(w, http.StatusNotFound, "invoice not found")
		return
	}
	writeJSON(w, http.StatusOK, stored)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /invoices", createInvoice)
	mux.HandleFunc("GET /invoices/{number}", getInvoice)

	http.ListenAndServe(":19067", mux)
}
