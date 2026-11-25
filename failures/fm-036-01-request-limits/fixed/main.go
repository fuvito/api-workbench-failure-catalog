package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Every documented limit is enforced, and breaking one gets a 4xx. The body
// size is limited while reading, so an oversized body is never held in memory.

const (
	maxBytes = 1024
	maxTags  = 10
	maxDepth = 5 // of the metadata object
)

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

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func depth(v any) int {
	deepest := 0
	switch value := v.(type) {
	case map[string]any:
		for _, child := range value {
			deepest = max(deepest, depth(child))
		}
	case []any:
		for _, child := range value {
			deepest = max(deepest, depth(child))
		}
	default:
		return 0
	}
	return deepest + 1
}

func createNote(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBytes))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, "body is larger than 1024 bytes")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid note")
		return
	}
	var n note
	if err := json.Unmarshal(body, &n); err != nil || n.Title == "" {
		writeError(w, http.StatusBadRequest, "invalid note")
		return
	}
	if len(n.Tags) > maxTags {
		writeError(w, http.StatusBadRequest, "at most 10 tags")
		return
	}
	if n.Metadata != nil && depth(n.Metadata) > maxDepth {
		writeError(w, http.StatusBadRequest, "metadata is nested more than 5 levels")
		return
	}
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
