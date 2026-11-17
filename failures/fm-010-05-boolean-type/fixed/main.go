package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type User struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Active        bool   `json:"active"`
	EmailVerified bool   `json:"email_verified"`
}

type FeatureFlags struct {
	DarkMode        bool `json:"dark_mode"`
	BetaFeatures    bool `json:"beta_features"`
	MaintenanceMode bool `json:"maintenance_mode"`
}

var users = map[int]User{
	1: {ID: 1, Name: "Alice", Active: true, EmailVerified: true},
	2: {ID: 2, Name: "Bob", Active: false, EmailVerified: true},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/users/"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}
		u, ok := users[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		}
		writeJSON(w, http.StatusOK, u)
	})

	mux.HandleFunc("/features", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, FeatureFlags{
			DarkMode:        true,
			BetaFeatures:    false,
			MaintenanceMode: false,
		})
	})

	http.ListenAndServe(":19029", mux)
}
