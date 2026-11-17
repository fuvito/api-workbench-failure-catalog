package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

// BUG: Post.Published is a pointer; when the client omits the field,
// it stays nil and marshals as null. The spec requires a boolean with
// a default of false — the server must apply that default on write.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type Post struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Published *bool  `json:"published"` // BUG: nil marshals as null
}

var (
	mu     sync.Mutex
	nextID = 1
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p Post
		json.NewDecoder(r.Body).Decode(&p)

		mu.Lock()
		p.ID = nextID
		nextID++
		mu.Unlock()

		// BUG: no default applied — Published remains nil if omitted by client
		writeJSON(w, http.StatusCreated, p)
	})

	http.ListenAndServe(":19019", mux)
}
