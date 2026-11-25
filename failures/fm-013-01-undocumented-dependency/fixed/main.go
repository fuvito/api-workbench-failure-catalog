package main

import (
	"encoding/json"
	"net/http"
)

// legacyAccountCode and the ACCOUNT_CLOSED code aren't in the public spec, but a
// registered consumer scenario relies on them, so they stay until it migrates.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func getAccount(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("id") == "acc-closed" {
		writeJSON(w, http.StatusGone, map[string]string{"code": "ACCOUNT_CLOSED", "message": "This account was closed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": "acc-1001", "name": "Northwind", "legacyAccountCode": "NW-0042"})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /accounts/{id}", getAccount)

	http.ListenAndServe(":19084", mux)
}
