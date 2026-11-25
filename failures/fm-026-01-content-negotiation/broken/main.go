package main

import (
	"encoding/json"
	"net/http"
)

// BUG: Accept is ignored, so every caller gets version 1 JSON with 200. The
// request Content-Type check compares the whole header, so a valid charset
// parameter is rejected.

func write(w http.ResponseWriter, status int, contentType string, v any) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	// BUG: Accept is never read
	write(w, http.StatusOK, "application/json", map[string]any{"id": "ord-1001", "total": 49.5})
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	// BUG: "application/json; charset=utf-8" is the same media type
	if r.Header.Get("Content-Type") != "application/json" {
		write(w, http.StatusUnsupportedMediaType, "application/json", map[string]string{"error": "unsupported media type"})
		return
	}
	var body struct {
		Total float64 `json:"total"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Total <= 0 {
		write(w, http.StatusBadRequest, "application/json", map[string]string{"error": "total is required"})
		return
	}
	write(w, http.StatusCreated, "application/json", map[string]any{"id": "ord-1002", "total": body.Total})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /orders/ord-1001", getOrder)
	mux.HandleFunc("POST /orders", createOrder)

	http.ListenAndServe(":19062", mux)
}
