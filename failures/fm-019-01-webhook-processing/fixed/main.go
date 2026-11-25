package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
)

// The signature is checked over the exact bytes received. Each event_id is
// processed once, and is remembered only after processing succeeds. A failure
// returns 500 so that the sender delivers the event again.

// Synthetic signing secret.
const secret = "demo-webhook-secret-not-real"

type event struct {
	EventID     string `json:"event_id"`
	Type        string `json:"type"`
	InvoiceID   string `json:"invoice_id"`
	AmountCents int    `json:"amount_cents"`
}

var (
	mu        sync.Mutex
	ledger    = []event{}
	processed = map[string]bool{}
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func signed(body []byte, header string) bool {
	encoded, ok := strings.CutPrefix(header, "sha256=")
	got, err := hex.DecodeString(encoded)
	if !ok || err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hmac.Equal(got, mac.Sum(nil))
}

func receive(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 64<<10))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid event"})
		return
	}
	// Verify before parsing or acting on anything in the body.
	if !signed(raw, r.Header.Get("X-Webhook-Signature")) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid signature"})
		return
	}
	var e event
	if json.Unmarshal(raw, &e) != nil || e.EventID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid event"})
		return
	}

	mu.Lock()
	defer mu.Unlock()
	if processed[e.EventID] {
		writeJSON(w, http.StatusOK, map[string]string{"result": "duplicate"})
		return
	}
	// Test hook: the ledger write "fails".
	if r.Header.Get("X-Test-Fault") == "ledger-write" {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "processing failed"})
		return
	}
	// In a real service, these two writes belong in one transaction.
	ledger = append(ledger, e)
	processed[e.EventID] = true
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
