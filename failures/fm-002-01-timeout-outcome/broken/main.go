package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// BUG: a slow request has no time limit, the client can't ask what happened to
// it, a retry creates a second order, and a cut-off response is announced as 200.

type order struct {
	ID    string `json:"id"`
	Total int    `json:"total_cents"`
}

var (
	mu     sync.Mutex
	nextID int
	byKey  = map[string][]order{}
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Idempotency-Key")
	mu.Lock()
	// BUG: every request creates a new order, even with the same key
	nextID++
	created := order{ID: fmt.Sprintf("ord-%d", nextID), Total: 4900}
	byKey[key] = append(byKey[key], created)
	mu.Unlock()

	// Test hook: the work after the commit is slow.
	if r.Header.Get("X-Test-Fault") == "slow-after-commit" {
		// BUG: the caller waits as long as it takes
		time.Sleep(3 * time.Second)
	}
	writeJSON(w, http.StatusCreated, created)
}

func countOrders(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]int{"count": len(byKey[r.URL.Query().Get("idempotency_key")])})
}

func export(w http.ResponseWriter, r *http.Request) {
	body := `{"orders":["ord-1","ord-2","ord-3"],"count":3}`
	w.Header().Set("Content-Type", "application/json")
	// Test hook: generating the export fails partway.
	if r.Header.Get("X-Test-Fault") == "export-fails-midway" {
		// BUG: 200 and the full length are already sent, then the body stops
		w.Header().Set("Content-Length", fmt.Sprint(len(body)))
		w.Write([]byte(body[:20]))
		return
	}
	w.Write([]byte(body))
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
