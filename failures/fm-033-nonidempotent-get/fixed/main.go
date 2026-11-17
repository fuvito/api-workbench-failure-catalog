package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

type Ticket struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

var (
	mu      sync.Mutex
	tickets = map[int]*Ticket{
		1: {ID: 1, Title: "Login broken", Status: "unread"},
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

	mux.HandleFunc("/tickets/", func(w http.ResponseWriter, r *http.Request) {
		id := parseID(r.URL.Path, "/tickets/")

		mu.Lock()
		defer mu.Unlock()

		t, ok := tickets[id]
		if !ok {
			http.NotFound(w, r)
			return
		}

		writeJSON(w, http.StatusOK, t)
	})

	http.ListenAndServe(":19033", mux)
}
