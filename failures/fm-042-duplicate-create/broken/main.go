package main

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

// BUG: creating a user with a duplicate email returns 201 and inserts
// a second record instead of returning 409 Conflict.

var nextID int64 = 1

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var users []User

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		// BUG: no duplicate check — always inserts
		u := User{ID: atomic.AddInt64(&nextID, 1) - 1, Name: body.Name, Email: body.Email}
		users = append(users, u)
		writeJSON(w, http.StatusCreated, u)
	})

	http.ListenAndServe(":19042", mux)
}
