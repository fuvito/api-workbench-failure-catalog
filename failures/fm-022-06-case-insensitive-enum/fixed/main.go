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

var validStatuses = map[string]bool{
	"pending":    true,
	"processing": true,
	"completed":  true,
	"cancelled":  true,
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)

		status, _ := body["status"].(string)
		if !validStatuses[status] {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{
				"error": "invalid status value",
			})
			return
		}

		writeJSON(w, http.StatusCreated, map[string]any{
			"id":     1,
			"status": status,
		})
	})

	http.ListenAndServe(":19037", mux)
}
