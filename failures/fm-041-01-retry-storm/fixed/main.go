package main

import (
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"
)

// Only temporary failures are retried, at most 3 attempts with growing, jittered
// delays. After a request uses all its attempts, the circuit opens and later
// requests fail at once, without calling inventory.

const (
	maxAttempts = 3
	baseDelay   = 20 * time.Millisecond
	openFor     = 30 * time.Second
)

var (
	mu        sync.Mutex
	mode      = "up" // of the simulated inventory service: up, failing or rejecting
	calls     int
	openUntil time.Time
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

func unavailable(w http.ResponseWriter) {
	w.Header().Set("Retry-After", "30")
	writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "inventory unavailable"})
}

func reserve(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	open := time.Now().Before(openUntil)
	mu.Unlock()
	if open {
		unavailable(w)
		return
	}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		switch callInventory() {
		case http.StatusOK:
			writeJSON(w, http.StatusCreated, map[string]string{"status": "reserved"})
			return
		case http.StatusUnprocessableEntity:
			// Retrying can't change a rejection.
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "reservation rejected"})
			return
		}
		if attempt < maxAttempts {
			delay := baseDelay << (attempt - 1)
			time.Sleep(delay/2 + rand.N(delay/2))
		}
	}
	mu.Lock()
	openUntil = time.Now().Add(openFor)
	mu.Unlock()
	unavailable(w)
}

// setInventory is a test hook. It also resets the call counter and the circuit.
func setInventory(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || (body.Mode != "up" && body.Mode != "failing" && body.Mode != "rejecting") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid mode"})
		return
	}
	mu.Lock()
	mode, calls, openUntil = body.Mode, 0, time.Time{}
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
