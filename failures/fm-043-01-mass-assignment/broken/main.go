package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

// BUG: PATCH decodes the request body straight into the stored customer record,
// so a caller can set any field the model has, including role, tenant_id and
// billing.credit_limit (mass assignment).

type billing struct {
	InvoiceEmail string `json:"invoice_email"`
	CreditLimit  int    `json:"credit_limit"`
}

type customer struct {
	ID       string  `json:"id"`
	TenantID string  `json:"tenant_id"`
	Name     string  `json:"name"`
	Role     string  `json:"role"`
	Billing  billing `json:"billing"`
}

// Synthetic bearer token for the one demo user.
const demoToken = "tok-alice"

var (
	mu     sync.Mutex
	record = customer{
		ID:       "cus-1001",
		TenantID: "tenant-a",
		Name:     "Alice",
		Role:     "member",
		Billing:  billing{InvoiceEmail: "alice@example.com", CreditLimit: 500},
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

func authenticated(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Authorization") != "Bearer "+demoToken {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}
	return true
}

func getMe(w http.ResponseWriter, r *http.Request) {
	if !authenticated(w, r) {
		return
	}
	mu.Lock()
	defer mu.Unlock()

	writeJSON(w, http.StatusOK, record)
}

func updateMe(w http.ResponseWriter, r *http.Request) {
	if !authenticated(w, r) {
		return
	}
	mu.Lock()
	defer mu.Unlock()

	// BUG: every JSON field the caller sends overwrites the stored one
	updated := record
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	record = updated
	writeJSON(w, http.StatusOK, record)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /customers/me", getMe)
	mux.HandleFunc("PATCH /customers/me", updateMe)

	http.ListenAndServe(":19052", mux)
}
