package main

import (
	"encoding/json"
	"net/http"
)

// BUG: the published lifecycle is not followed. /v1/invoices is gone before its
// sunset date, /v1/customers gives no sign that it is deprecated, and the
// retired /v0/invoices answers like a mistyped URL.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func list(id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"items": []map[string]string{{"id": id}}})
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /v2/invoices", list("inv-1001"))
	mux.HandleFunc("GET /v2/customers", list("cus-1001"))

	// BUG: removed early
	mux.HandleFunc("GET /v1/invoices", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	})
	// BUG: no Deprecation, Sunset or Link headers
	mux.HandleFunc("GET /v1/customers", list("cus-1001"))
	// BUG: a retired endpoint should say it is gone
	mux.HandleFunc("GET /v0/invoices", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	})

	http.ListenAndServe(":19060", mux)
}
