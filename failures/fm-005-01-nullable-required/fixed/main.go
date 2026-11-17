package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"` // always a non-null string
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

var users = map[int]User{
	1: {ID: 1, Name: "Alice Nguyen", Email: "alice@example.com"},
	2: {ID: 2, Name: "Bob Chen", Email: ""}, // OAuth user — empty string satisfies type: string
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

		writeJSON(w, http.StatusOK, user)
	})

	log.Fatal(http.ListenAndServe(":19005", mux))
}
