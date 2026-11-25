package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"
)

// BUG: reading an order waits for the optional recommendation service and fails
// when that service fails.

var (
	mu    sync.Mutex
	state = "up" // of the simulated recommendation service: up, down or slow
)

// recommendations stands in for a call to another service.
func recommendations() ([]string, error) {
	mu.Lock()
	current := state
	mu.Unlock()

	switch current {
	case "down":
		return nil, errors.New("connection refused")
	case "slow":
		time.Sleep(3 * time.Second)
	}
	return []string{"lamp-shade"}, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	// BUG: no time limit, and a failure here fails the whole request
	items, err := recommendations()
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "service unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": "ord-1001", "total": 49.5, "recommendations": items, "degraded": false})
}

func getRecommendations(w http.ResponseWriter, r *http.Request) {
	items, err := recommendations()
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "service unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"recommendations": items})
}

// setState is a test hook that switches the simulated dependency.
func setState(w http.ResponseWriter, r *http.Request) {
	var body struct {
		State string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || (body.State != "up" && body.State != "down" && body.State != "slow") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "state must be up, down or slow"})
		return
	}
	mu.Lock()
	state = body.State
	mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /orders/ord-1001", getOrder)
	mux.HandleFunc("GET /orders/ord-1001/recommendations", getRecommendations)
	mux.HandleFunc("PUT /test/recommendation-service", setState)

	http.ListenAndServe(":19066", mux)
}
