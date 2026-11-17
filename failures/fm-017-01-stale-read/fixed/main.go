package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var (
	mu    sync.Mutex
	items = map[int]*Item{
		1: {ID: 1, Name: "Original"},
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
		if id == 0 {
			http.NotFound(w, r)
			return
		}

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

		case http.MethodPut:
			var updated Item
			json.NewDecoder(r.Body).Decode(&updated)
			updated.ID = id
			items[id] = &updated
			writeJSON(w, http.StatusOK, items[id])

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.ListenAndServe(":19017", mux)
}
