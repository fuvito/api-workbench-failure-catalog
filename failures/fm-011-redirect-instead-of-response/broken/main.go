package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// BUG: GET /posts/{id} returns 301 to a legacy URL instead of the post.
// A middleware was added to redirect old-format IDs but fires for all requests.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type Post struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

var posts = map[int]Post{
	1: {ID: 1, Title: "Hello World", Body: "First post."},
	2: {ID: 2, Title: "API Design", Body: "Best practices."},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/posts/", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/posts/"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}
		// BUG: redirect middleware fires unconditionally
		http.Redirect(w, r, "/posts/legacy/"+strconv.Itoa(id), http.StatusMovedPermanently)
	})

	http.ListenAndServe(":19011", mux)
}
