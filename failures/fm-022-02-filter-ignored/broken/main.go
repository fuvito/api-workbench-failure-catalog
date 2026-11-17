package main

import (
	"encoding/json"
	"net/http"
)

// BUG: the status filter query parameter is accepted but ignored.
// All orders are always returned regardless of ?status=.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type Order struct {
	ID     int     `json:"id"`
	Status string  `json:"status"`
	Total  float64 `json:"total"`
}

var orders = []Order{
	{ID: 1, Status: "pending", Total: 29.99},
	{ID: 2, Status: "completed", Total: 59.99},
	{ID: 3, Status: "completed", Total: 14.99},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		// BUG: filter ignored — always returns all orders
		_ = r.URL.Query().Get("status")
		writeJSON(w, http.StatusOK, map[string]any{"items": orders, "count": len(orders)})
	})

	http.ListenAndServe(":19028", mux)
}
