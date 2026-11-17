package main

import (
	"encoding/json"
	"net/http"
)

// BUG: GET /users returns a bare JSON array instead of the declared object wrapper
// {"items":[...],"count":N}. The array type fails the response schema's type: object.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var users = []User{
	{ID: 1, Name: "Alice", Email: "alice@example.com"},
	{ID: 2, Name: "Bob", Email: "bob@example.com"},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		// BUG: returns bare array instead of {items:[...], count:N}
		writeJSON(w, http.StatusOK, users)
	})

	http.ListenAndServe(":19008", mux)
}
