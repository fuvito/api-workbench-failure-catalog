package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"unicode/utf8"
)

// Names are stored exactly as sent. The limit counts characters (Unicode code
// points), as maxLength does in the spec, and a longer name is rejected.

const maxName = 20

type contact struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var (
	mu       sync.Mutex
	contacts = map[string]contact{}
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func createContact(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if utf8.RuneCountInString(body.Name) > maxName {
		writeError(w, http.StatusBadRequest, "name must be at most 20 characters")
		return
	}
	mu.Lock()
	defer mu.Unlock()

	c := contact{ID: fmt.Sprintf("con-%d", len(contacts)+1), Name: body.Name}
	contacts[c.ID] = c
	writeJSON(w, http.StatusCreated, c)
}

func getContact(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	c, ok := contacts[r.PathValue("id")]
	if !ok {
		writeError(w, http.StatusNotFound, "contact not found")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /contacts", createContact)
	mux.HandleFunc("GET /contacts/{id}", getContact)

	http.ListenAndServe(":19055", mux)
}
