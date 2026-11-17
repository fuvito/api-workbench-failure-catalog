package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// BUG: boolean fields are returned as strings ("true"/"false") for users
// and as integers (1/0) for feature flags, instead of JSON booleans.
// This happens when values come from a database driver that maps BIT or
// TINYINT columns to int64, or a serializer that quotes all scalar values.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type UserBroken struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Active        string `json:"active"`        // BUG: "true" / "false" (string)
	EmailVerified string `json:"email_verified"` // BUG: "true" / "false" (string)
}

type FeatureFlagsBroken struct {
	DarkMode        int `json:"dark_mode"`        // BUG: 1 / 0 (integer)
	BetaFeatures    int `json:"beta_features"`    // BUG: 1 / 0 (integer)
	MaintenanceMode int `json:"maintenance_mode"` // BUG: 1 / 0 (integer)
}

var users = map[int]UserBroken{
	1: {ID: 1, Name: "Alice", Active: "true", EmailVerified: "true"},
	2: {ID: 2, Name: "Bob", Active: "false", EmailVerified: "true"},
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
		writeJSON(w, http.StatusOK, FeatureFlagsBroken{
			DarkMode:        1,
			BetaFeatures:    0,
			MaintenanceMode: 0,
		})
	})

	http.ListenAndServe(":19029", mux)
}
