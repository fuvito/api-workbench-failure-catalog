package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// BUG: struct uses camelCase JSON tags instead of the snake_case field names
// declared in the OpenAPI spec. Required fields like user_id and first_name
// are absent; additionalProperties: false rejects the undeclared camelCase names.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type UserBroken struct {
	UserID    int    `json:"userId"`    // BUG: should be user_id
	FirstName string `json:"firstName"` // BUG: should be first_name
	LastName  string `json:"lastName"`  // BUG: should be last_name
	Email     string `json:"email"`
}

func parseID(path, prefix string) int {
	s := strings.TrimPrefix(path, prefix)
	id := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			id = id*10 + int(c-'0')
		}
	}
	return id
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		id := parseID(r.URL.Path, "/users/")
		if id != 1 {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, UserBroken{
			UserID:    1,
			FirstName: "Alice",
			LastName:  "Smith",
			Email:     "alice@example.com",
		})
	})

	http.ListenAndServe(":19009", mux)
}
