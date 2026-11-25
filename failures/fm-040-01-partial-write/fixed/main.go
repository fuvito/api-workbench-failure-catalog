package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
)

// The whole invoice is built first and stored with one write, only after every
// line has succeeded. A real service does this with a database transaction that
// rolls back on any error.

type line struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

type invoice struct {
	Number string `json:"number"`
	Lines  []line `json:"lines"`
}

var (
	errWrite     = errors.New("write failed")
	errDuplicate = errors.New("duplicate sku")
)

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

// stage prepares the invoice without storing anything.
func stage(in invoice, failSecondLine bool) (*invoice, error) {
	staged := &invoice{Number: in.Number, Lines: []line{}}
	seen := map[string]bool{}
	for i, l := range in.Lines {
		if i == 1 && failSecondLine {
			return nil, errWrite
		}
		// A unique constraint: one line per SKU.
		if seen[l.SKU] {
			return nil, errDuplicate
		}
		seen[l.SKU] = true
		staged.Lines = append(staged.Lines, l)
	}
	return staged, nil
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

	staged, err := stage(in, failSecondLine)
	switch {
	case errors.Is(err, errDuplicate):
		writeError(w, http.StatusConflict, "duplicate sku")
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	invoices[staged.Number] = staged // commit
	writeJSON(w, http.StatusCreated, staged)
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
