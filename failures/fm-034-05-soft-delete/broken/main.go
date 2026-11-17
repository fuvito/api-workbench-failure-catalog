package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// BUG: DELETE permanently removes the item (hard delete) and returns 204 No Content.
// The spec requires a soft delete: return 200 with the item and a deleted_at timestamp.

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var (
	mu    sync.Mutex
	items = map[int]*Item{
		1: {ID: 1, Name: "Widget"},
	}
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
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

	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		id := parseID(r.URL.Path, "/items/")

		mu.Lock()
		defer mu.Unlock()

		switch r.Method {
		case http.MethodGet:
			item, ok := items[id]
			if !ok {
				http.NotFound(w, r)
				return
			}
			writeJSON(w, http.StatusOK, item)

		case http.MethodDelete:
			_, ok := items[id]
			if !ok {
				http.NotFound(w, r)
				return
			}
			// BUG: hard delete — removes item from store and returns 204
			delete(items, id)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.ListenAndServe(":19036", mux)
}
