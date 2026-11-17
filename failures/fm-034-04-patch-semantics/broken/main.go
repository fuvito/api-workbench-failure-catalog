package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// BUG: PATCH behaves like PUT — it replaces the entire user record with
// only the fields provided in the request body, zeroing out all others.
// A PATCH with {"email":"new@example.com"} should leave name and role
// unchanged, but here it wipes them.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

var users = map[int]*User{
	1: {ID: 1, Name: "Alice", Email: "alice@example.com", Role: "admin"},
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

		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, u)

		case http.MethodPatch:
			var patch User
			json.NewDecoder(r.Body).Decode(&patch)
			// BUG: replaces the whole record instead of merging
			patch.ID = id
			users[id] = &patch
			writeJSON(w, http.StatusOK, users[id])

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	http.ListenAndServe(":19038", mux)
}
