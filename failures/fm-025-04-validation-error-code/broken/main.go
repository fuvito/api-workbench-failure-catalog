package main

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

// BUG: when required fields are missing the handler panics or returns
// 500 Internal Server Error instead of 422 Unprocessable Entity.

var nextID int64 = 1

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Name  string  `json:"name"`
			Price float64 `json:"price"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			// BUG: returns 500 for any decode/validation error
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
		if body.Name == "" || body.Price == 0 {
			// BUG: returns 500 instead of 422
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
		id := atomic.AddInt64(&nextID, 1) - 1
		writeJSON(w, http.StatusCreated, map[string]any{"id": id, "name": body.Name, "price": body.Price})
	})

	http.ListenAndServe(":19043", mux)
}
