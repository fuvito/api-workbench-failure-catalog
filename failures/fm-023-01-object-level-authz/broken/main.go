package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// BUG: every invoice handler checks that the caller is authenticated, but never
// that the caller's tenant owns the invoice named in the URL. Changing the ID
// exposes another tenant's records (broken object-level authorization).

type invoice struct {
	ID       string  `json:"id"`
	TenantID string  `json:"tenant_id"`
	Amount   float64 `json:"amount"`
	Memo     string  `json:"memo"`
}

// Synthetic bearer tokens, one user per tenant.
var tenantByToken = map[string]string{
	"tok-alice": "tenant-a",
	"tok-bob":   "tenant-b",
}

var (
	mu       sync.Mutex
	invoices = map[string]invoice{
		"inv-1001": {ID: "inv-1001", TenantID: "tenant-a", Amount: 120.50, Memo: "Alice consulting"},
		"inv-1002": {ID: "inv-1002", TenantID: "tenant-b", Amount: 980.00, Memo: "Bob hardware"},
	}
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func authenticate(w http.ResponseWriter, r *http.Request) (string, bool) {
	token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	tenant, known := tenantByToken[token]
	if !found || !known {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return "", false
	}
	return tenant, true
}

func getInvoice(w http.ResponseWriter, r *http.Request) {
	if _, ok := authenticate(w, r); !ok {
		return
	}
	mu.Lock()
	defer mu.Unlock()

	// BUG: looks the invoice up by ID alone
	inv, ok := invoices[r.PathValue("id")]
	if !ok {
		writeError(w, http.StatusNotFound, "invoice not found")
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

func updateInvoice(w http.ResponseWriter, r *http.Request) {
	if _, ok := authenticate(w, r); !ok {
		return
	}
	var body struct {
		Memo string `json:"memo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Memo == "" {
		writeError(w, http.StatusBadRequest, "memo is required")
		return
	}
	mu.Lock()
	defer mu.Unlock()

	// BUG: any authenticated caller can change any invoice
	inv, ok := invoices[r.PathValue("id")]
	if !ok {
		writeError(w, http.StatusNotFound, "invoice not found")
		return
	}
	inv.Memo = body.Memo
	invoices[inv.ID] = inv
	writeJSON(w, http.StatusOK, inv)
}

func deleteInvoice(w http.ResponseWriter, r *http.Request) {
	if _, ok := authenticate(w, r); !ok {
		return
	}
	mu.Lock()
	defer mu.Unlock()

	// BUG: any authenticated caller can delete any invoice
	inv, ok := invoices[r.PathValue("id")]
	if !ok {
		writeError(w, http.StatusNotFound, "invoice not found")
		return
	}
	delete(invoices, inv.ID)
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /invoices/{id}", getInvoice)
	mux.HandleFunc("PATCH /invoices/{id}", updateInvoice)
	mux.HandleFunc("DELETE /invoices/{id}", deleteInvoice)

	http.ListenAndServe(":19051", mux)
}
