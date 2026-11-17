package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// BUG: when a team has no members or a project has no tags, the server
// returns null for the collection field instead of an empty array [].
// In Go, a nil slice marshals to null; it must be initialized to []T{}.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type Member struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type MemberList struct {
	Members []Member `json:"members"` // BUG: nil slice → marshals to null
}

type TagList struct {
	Tags []string `json:"tags"` // BUG: nil slice → marshals to null
}

var teamMembers = map[int][]Member{
	1: {{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}},
	2: nil, // empty team — BUG: nil marshals to null
}

var projectTags = map[int][]string{
	1: {"backend", "go"},
	2: nil, // no tags — BUG: nil marshals to null
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/teams/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 3 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}
		members, ok := teamMembers[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "team not found"})
			return
		}
		writeJSON(w, http.StatusOK, MemberList{Members: members})
	})

	mux.HandleFunc("/projects/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 3 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}
		tags, ok := projectTags[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
			return
		}
		writeJSON(w, http.StatusOK, TagList{Tags: tags})
	})

	http.ListenAndServe(":19023", mux)
}
