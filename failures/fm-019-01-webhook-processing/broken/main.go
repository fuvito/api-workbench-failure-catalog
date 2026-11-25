package main

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

// BUG: the webhook consumer doesn't verify signatures, processes every delivery
// as new, and acknowledges an event even when processing it failed.

type event struct {
	EventID     string `json:"event_id"`
	Type        string `json:"type"`
	InvoiceID   string `json:"invoice_id"`
	AmountCents int    `json:"amount_cents"`
}

var (
	mu     sync.Mutex
	ledger = []event{}
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func receive(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(r.Body)
	var e event
	if err != nil || json.Unmarshal(raw, &e) != nil || e.EventID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid event"})
		return
	}
	// BUG: X-Webhook-Signature is never checked

	// Test hook: the ledger write "fails".
	if r.Header.Get("X-Test-Fault") == "ledger-write" {
		// BUG: tells the sender it succeeded, so the event is never sent again
		writeJSON(w, http.StatusOK, map[string]string{"result": "processed"})
		return
	}

	mu.Lock()
	defer mu.Unlock()
	// BUG: no check for an event that was already processed
	ledger = append(ledger, e)
	writeJSON(w, http.StatusOK, map[string]string{"result": "processed"})
}

func getLedger(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"count": len(ledger), "entries": ledger})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /webhooks/billing", receive)
	mux.HandleFunc("GET /ledger", getLedger)

	http.ListenAndServe(":19070", mux)
}
