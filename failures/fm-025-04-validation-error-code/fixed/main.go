package main

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

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
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
				"error":  "invalid request body",
				"fields": []string{"body"},
			})
			return
		}
		var missing []string
		if body.Name == "" {
			missing = append(missing, "name")
		}
		if body.Price == 0 {
			missing = append(missing, "price")
		}
		if len(missing) > 0 {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
				"error":  "missing required fields",
				"fields": missing,
			})
			return
		}
		id := atomic.AddInt64(&nextID, 1) - 1
		writeJSON(w, http.StatusCreated, map[string]any{"id": id, "name": body.Name, "price": body.Price})
	})

	http.ListenAndServe(":19043", mux)
}
