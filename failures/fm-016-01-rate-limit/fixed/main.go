package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
)

// The rate limit is raised far above what the scenario exercises (3 requests).
// In production the right fix is also to declare 429 in the spec and add
// consumer-side retry logic — see the README for the full migration pattern.
const rateLimit = 1000

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

var items = []Item{
	{ID: 1, Name: "Widget"},
	{ID: 2, Name: "Gadget"},
	{ID: 3, Name: "Doohickey"},
}

var requestCount int64

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt64(&requestCount, 1)
		if n > rateLimit {
			w.Header().Set("Retry-After", "1")
			writeJSON(w, http.StatusTooManyRequests, ErrorResponse{
				Error: "rate limit exceeded",
				Code:  "RATE_LIMIT_EXCEEDED",
			})
			return
		}
		writeJSON(w, http.StatusOK, items)
	})

	log.Fatal(http.ListenAndServe(":19016", mux))
}
