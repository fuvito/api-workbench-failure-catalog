package main

import (
	"encoding/json"
	"net/http"
)

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
		statusFilter := r.URL.Query().Get("status")
		result := []Order{}
		for _, o := range orders {
			if statusFilter == "" || o.Status == statusFilter {
				result = append(result, o)
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": result, "count": len(result)})
	})

	http.ListenAndServe(":19028", mux)
}
