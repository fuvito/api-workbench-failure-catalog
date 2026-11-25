package main

import (
	"encoding/json"
	"net/http"
)

// BUG: legacyAccountCode and the ACCOUNT_CLOSED code were never documented, so
// a cleanup removed them. A finance export job still reads both.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func getAccount(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("id") == "acc-closed" {
		// BUG: the code the finance job branches on is gone
		writeJSON(w, http.StatusGone, map[string]string{"message": "This account was closed"})
		return
	}
	// BUG: legacyAccountCode removed as "unused"
	writeJSON(w, http.StatusOK, map[string]string{"id": "acc-1001", "name": "Northwind"})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /accounts/{id}", getAccount)

	http.ListenAndServe(":19084", mux)
}
