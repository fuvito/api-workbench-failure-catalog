package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type User struct {
	ID           int    `json:"-"`
	Name         string `json:"-"`
	Email        string `json:"-"`
	PasswordHash string `json:"-"`
	APIKey       string `json:"-"`
	Role         string `json:"-"`
}

// UserResponse is the shape declared in openapi.yaml — only safe, public fields.
type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
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
		writeJSON(w, http.StatusOK, UserResponse{ID: user.ID, Name: user.Name, Email: user.Email})
	})

	log.Fatal(http.ListenAndServe(":19044", mux))
}
