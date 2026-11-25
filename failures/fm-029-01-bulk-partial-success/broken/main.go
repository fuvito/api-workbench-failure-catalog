package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// BUG: the batch endpoint reports only how many items succeeded and failed, not
// which ones. It also processes every item it receives, so a client that sends
// the batch again pays the successful items a second time.

type item struct {
	Ref         string `json:"ref"`
	AmountCents int    `json:"amount_cents"`
}

type payment struct {
	ID          string `json:"id"`
	Ref         string `json:"ref"`
	AmountCents int    `json:"amount_cents"`
}

var (
	mu       sync.Mutex
	payments = []payment{}
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func createBatch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Items []item `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Items) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "items is required"})
		return
	}
	mu.Lock()
	defer mu.Unlock()

	succeeded, failed := 0, 0
	for _, it := range body.Items {
		if it.AmountCents <= 0 {
			failed++
			continue
		}
		// BUG: no check whether this ref was already paid
		payments = append(payments, payment{
			ID:          fmt.Sprintf("pmt-%d", len(payments)+1),
			Ref:         it.Ref,
			AmountCents: it.AmountCents,
		})
		succeeded++
	}
	// BUG: no per-item results
	writeJSON(w, http.StatusOK, map[string]int{"succeeded": succeeded, "failed": failed})
}

func listPayments(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{"count": len(payments), "payments": payments})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /payments/batch", createBatch)
	mux.HandleFunc("GET /payments", listPayments)

	http.ListenAndServe(":19053", mux)
}
