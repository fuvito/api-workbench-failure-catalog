package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// BUG: enum validation is case-insensitive; "PENDING" (uppercase) is accepted
// when only "pending" (lowercase) is declared in the spec.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

var validStatuses = []string{"pending", "processing", "completed", "cancelled"}

func isValidStatus(s string) bool {
	// BUG: lowercases the input before comparing, making validation case-insensitive
	lower := strings.ToLower(s)
	for _, v := range validStatuses {
		if v == lower {
			return true
		}
	}
	return false
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
		if !isValidStatus(status) {
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
