package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// BUG: the full internal User struct is serialised directly into the response.
// Fields password_hash, api_key, and role are not declared in openapi.yaml,
// which uses additionalProperties: false. The specFile: binding catches them
// as undeclared properties and exits 1.

type User struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"` // sensitive — must not be exposed
	APIKey       string `json:"api_key"`        // sensitive — must not be exposed
	Role         string `json:"role"`           // internal — must not be exposed
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

var users = map[int]User{
	1: {
		ID:           1,
		Name:         "Alice Nguyen",
		Email:        "alice@example.com",
		PasswordHash: "$2b$12$KIXabc123hashedvalue",
		APIKey:       "sk_live_abc123def456",
		Role:         "admin",
	},
	2: {
		ID:           2,
		Name:         "Bob Chen",
		Email:        "bob@example.com",
		PasswordHash: "$2b$12$KIXxyz789hashedvalue",
		APIKey:       "sk_live_xyz789uvw012",
		Role:         "member",
	},
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid id", Code: "INVALID_ID"})
			return
		}
		user, ok := users[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not found", Code: "USER_NOT_FOUND"})
			return
		}
		writeJSON(w, http.StatusOK, user) // BUG: exposes password_hash, api_key, role
	})

	log.Fatal(http.ListenAndServe(":19044", mux))
}
