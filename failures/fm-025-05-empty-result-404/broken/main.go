package main

import (
	"encoding/json"
	"net/http"
)

// BUG: when a filter query returns no results, the server returns 404 Not Found
// instead of 200 OK with an empty list.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

var users = []User{
	{ID: 1, Name: "Alice", Role: "admin"},
	{ID: 2, Name: "Bob", Role: "viewer"},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		role := r.URL.Query().Get("role")
		var result []User
		for _, u := range users {
			if role == "" || u.Role == role {
				result = append(result, u)
			}
		}

		// BUG: treats an empty result set as "not found"
		if len(result) == 0 {
			http.NotFound(w, r)
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"items": result,
			"count": len(result),
		})
	})

	http.ListenAndServe(":19048", mux)
}
