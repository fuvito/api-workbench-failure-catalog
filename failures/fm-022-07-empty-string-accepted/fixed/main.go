package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

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

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			Username string `json:"username"`
			Email    string `json:"email"`
		}
		json.NewDecoder(r.Body).Decode(&body)

		if strings.TrimSpace(body.Username) == "" || strings.TrimSpace(body.Email) == "" {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{
				"error": "username and email must not be empty",
			})
			return
		}

		writeJSON(w, http.StatusCreated, map[string]any{
			"id":       1,
			"username": body.Username,
			"email":    body.Email,
		})
	})

	http.ListenAndServe(":19045", mux)
}
