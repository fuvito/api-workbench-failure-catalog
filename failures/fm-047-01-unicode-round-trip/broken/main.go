package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"unicode"
)

// BUG: names are damaged before they are stored. Combining marks are dropped,
// characters above U+FFFF become "?", and the 20-character limit is applied by
// cutting the string at 20 bytes.

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

func clean(name string) string {
	name = strings.Map(func(r rune) rune {
		// BUG: an "input sanitizer" that removes accents sent as combining marks
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		// BUG: behaves like storage that holds at most 3 bytes per character
		if r > 0xFFFF {
			return '?'
		}
		return r
	}, name)

	// BUG: len counts bytes, so this can cut a character in half. It also
	// truncates silently instead of rejecting the request.
	if len(name) > maxName {
		name = name[:maxName]
	}
	return name
}

func createContact(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	mu.Lock()
	defer mu.Unlock()

	c := contact{ID: fmt.Sprintf("con-%d", len(contacts)+1), Name: clean(body.Name)}
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
