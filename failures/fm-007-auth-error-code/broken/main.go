package main

import (
	"encoding/json"
	"net/http"
)

// BUG: auth failures return 200 with an error body instead of 401.
// Many frameworks/middlewares do this when "success" is determined by
// checking response.ok in client code rather than the HTTP status code.

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

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		json.NewDecoder(r.Body).Decode(&body)

		if body.Username == "alice" && body.Password == "secret" {
			writeJSON(w, http.StatusOK, map[string]string{"token": "tok-abc123"})
			return
		}
		// BUG: returns 200 instead of 401
		writeJSON(w, http.StatusOK, map[string]string{"error": "invalid credentials"})
	})

	mux.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Authorization") != "Bearer tok-abc123" {
			// BUG: returns 200 instead of 401
			writeJSON(w, http.StatusOK, map[string]string{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"id": 1, "username": "alice"})
	})

	http.ListenAndServe(":19007", mux)
}
