package main

import (
	"encoding/json"
	"net/http"
)

// Deprecated endpoints keep working until their sunset date and say so in
// headers. Retired endpoints answer 410 and name their replacement.

const (
	// Deprecation (RFC 9745) is a Unix time: 2026-06-01T00:00:00Z.
	deprecatedAt = "@1780272000"
	// Sunset (RFC 8594) is an HTTP date.
	sunsetAt = "Fri, 31 Dec 2027 23:59:59 GMT"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func successor(path string) string { return "<" + path + `>; rel="successor-version"` }

func list(id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"items": []map[string]string{{"id": id}}})
	}
}

func deprecated(replacement string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Deprecation", deprecatedAt)
		w.Header().Set("Sunset", sunsetAt)
		w.Header().Set("Link", successor(replacement))
		next(w, r)
	}
}

func retired(replacement string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Link", successor(replacement))
		writeJSON(w, http.StatusGone, map[string]string{"error": "retired, use " + replacement})
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /v2/invoices", list("inv-1001"))
	mux.HandleFunc("GET /v2/customers", list("cus-1001"))
	mux.HandleFunc("GET /v1/invoices", deprecated("/v2/invoices", list("inv-1001")))
	mux.HandleFunc("GET /v1/customers", deprecated("/v2/customers", list("cus-1001")))
	mux.HandleFunc("GET /v0/invoices", retired("/v2/invoices"))

	http.ListenAndServe(":19060", mux)
}
