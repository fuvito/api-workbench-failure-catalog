package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// BUG: created_at is returned as a Unix timestamp integer instead of an ISO 8601 string.
// scheduled_for is returned in MM/DD/YYYY format instead of YYYY-MM-DD.
// Both violate the declared OpenAPI schema types.

type EventBroken struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	CreatedAt    int64  `json:"created_at"`    // BUG: should be string (ISO 8601)
	ScheduledFor string `json:"scheduled_for"` // BUG: MM/DD/YYYY instead of YYYY-MM-DD
}

var events = []EventBroken{
	{ID: 1, Name: "API Summit", CreatedAt: 1731196800, ScheduledFor: "12/15/2026"},
	{ID: 2, Name: "Platform Review", CreatedAt: 1731283200, ScheduledFor: "01/08/2027"},
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/events/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		idStr := strings.TrimPrefix(r.URL.Path, "/events/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}
		for _, e := range events {
			if e.ID == id {
				writeJSON(w, http.StatusOK, e)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	})

	http.ListenAndServe(":19002", mux)
}
