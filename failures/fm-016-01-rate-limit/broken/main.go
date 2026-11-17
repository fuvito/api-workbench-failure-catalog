package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
)

// BUG: the server enforces a rate limit and returns 429, but 429 is not declared
// in the OpenAPI spec. Consumer code has no contract telling it to handle throttling.
const rateLimit = 2 // requests allowed before throttling kicks in

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
			// BUG: 429 returned but not declared in openapi.yaml —
			// clrsig catches this as an undeclared response status
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
