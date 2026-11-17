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
			// Decode into a map so we only touch fields that were actually sent
			var patch map[string]string
			json.NewDecoder(r.Body).Decode(&patch)
			if v, ok := patch["name"]; ok {
				u.Name = v
			}
			if v, ok := patch["email"]; ok {
				u.Email = v
			}
			writeJSON(w, http.StatusOK, u)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	http.ListenAndServe(":19038", mux)
}
