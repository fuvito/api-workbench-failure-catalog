package main

import (
	"encoding/json"
	"net/http"
)

// BUG: when status is "shipped", tracking_number is required by the spec but
// the server never enforces this conditional constraint.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

var validStatuses = map[string]bool{
	"pending":    true,
	"processing": true,
	"shipped":    true,
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/shipments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			Status         string `json:"status"`
			TrackingNumber string `json:"tracking_number"`
		}
		json.NewDecoder(r.Body).Decode(&body)

		if !validStatuses[body.Status] {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{
				"error": "invalid status",
			})
			return
		}

		// BUG: does not enforce tracking_number requirement when status == "shipped"
		writeJSON(w, http.StatusCreated, map[string]any{
			"id":              1,
			"status":          body.Status,
			"tracking_number": body.TrackingNumber,
		})
	})

	http.ListenAndServe(":19047", mux)
}
