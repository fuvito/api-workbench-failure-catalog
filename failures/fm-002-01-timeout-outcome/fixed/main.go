package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// A request that can't finish within the budget gets 504 saying the outcome is
// unknown. Retrying with the same Idempotency-Key returns the order that was
// already created. The export is built completely before anything is sent.

const budget = 500 * time.Millisecond

type order struct {
	ID    string `json:"id"`
	Total int    `json:"total_cents"`
}

var (
	mu     sync.Mutex
	nextID int
	byKey  = map[string]order{}
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Idempotency-Key is required"})
		return
	}
	mu.Lock()
	existing, seen := byKey[key]
	if !seen {
		nextID++
		existing = order{ID: fmt.Sprintf("ord-%d", nextID), Total: 4900}
		byKey[key] = existing
	}
	mu.Unlock()
	if seen {
		writeJSON(w, http.StatusOK, existing)
		return
	}

	// Test hook: the work after the commit is slow.
	delay := time.Duration(0)
	if r.Header.Get("X-Test-Fault") == "slow-after-commit" {
		delay = 3 * time.Second
	}
	select {
	case <-time.After(delay):
		writeJSON(w, http.StatusCreated, existing)
	case <-time.After(budget):
		writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "timed out", "outcome": "unknown"})
	}
}

func countOrders(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	count := 0
	if _, ok := byKey[r.URL.Query().Get("idempotency_key")]; ok {
		count = 1
	}
	writeJSON(w, http.StatusOK, map[string]int{"count": count})
}

func export(w http.ResponseWriter, r *http.Request) {
	// Test hook: generating the export fails partway. Nothing has been sent yet,
	// so the failure can still be reported properly.
	if r.Header.Get("X-Test-Fault") == "export-fails-midway" {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "export failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"orders": []string{"ord-1", "ord-2", "ord-3"}, "count": 3})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /orders", createOrder)
	mux.HandleFunc("GET /orders/count", countOrders)
	mux.HandleFunc("GET /orders/export", export)

	http.ListenAndServe(":19082", mux)
}
