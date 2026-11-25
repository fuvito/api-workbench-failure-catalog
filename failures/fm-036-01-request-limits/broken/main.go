package main

import (
	"encoding/json"
	"io"
	"net/http"
)

// BUG: the documented limits are not enforced. An oversized body causes a 500,
// and there is no limit on the number of tags or on nesting depth.

// parserLimit stands for a smaller limit hidden inside the backend.
const parserLimit = 1024

type note struct {
	Title    string         `json:"title"`
	Tags     []string       `json:"tags"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func createNote(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	if len(body) > parserLimit {
		// BUG: a client's oversized request is reported as a server fault
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "parser buffer exceeded"})
		return
	}
	var n note
	if err := json.Unmarshal(body, &n); err != nil || n.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid note"})
		return
	}
	// BUG: tags and nesting depth are not checked
	writeJSON(w, http.StatusCreated, map[string]any{"title": n.Title, "tag_count": len(n.Tags)})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /notes", createNote)

	http.ListenAndServe(":19063", mux)
}
