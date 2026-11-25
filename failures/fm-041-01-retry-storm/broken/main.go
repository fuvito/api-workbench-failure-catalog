package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

// BUG: every failed call to inventory is retried at once, up to 10 attempts,
// even when inventory rejected the request, and each new request does it again.

const maxAttempts = 10

var (
	mu    sync.Mutex
	mode  = "up" // of the simulated inventory service: up, failing or rejecting
	calls int
)

// callInventory stands in for a call to another service and returns its status.
func callInventory() int {
	mu.Lock()
	defer mu.Unlock()
	calls++
	switch mode {
	case "failing":
		return http.StatusServiceUnavailable
	case "rejecting":
		return http.StatusUnprocessableEntity
	}
	return http.StatusOK
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func reserve(w http.ResponseWriter, r *http.Request) {
	status := 0
	// BUG: no backoff, no distinction between errors, no circuit breaker
	for attempt := 0; attempt < maxAttempts && status != http.StatusOK; attempt++ {
		status = callInventory()
	}
	if status != http.StatusOK {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "inventory unavailable"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "reserved"})
}

// setInventory is a test hook. It also resets the call counter.
func setInventory(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || (body.Mode != "up" && body.Mode != "failing" && body.Mode != "rejecting") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid mode"})
		return
	}
	mu.Lock()
	mode, calls = body.Mode, 0
	mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func inventoryStats(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]int{"calls": calls})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /reservations", reserve)
	mux.HandleFunc("PUT /test/inventory", setInventory)
	mux.HandleFunc("GET /test/inventory", inventoryStats)

	http.ListenAndServe(":19068", mux)
}
