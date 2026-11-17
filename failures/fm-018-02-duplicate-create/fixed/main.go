package main

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

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
		for _, u := range users {
			if u.Email == body.Email {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "email already exists"})
				return
			}
		}
		u := User{ID: atomic.AddInt64(&nextID, 1) - 1, Name: body.Name, Email: body.Email}
		users = append(users, u)
		writeJSON(w, http.StatusCreated, u)
	})

	http.ListenAndServe(":19042", mux)
}
