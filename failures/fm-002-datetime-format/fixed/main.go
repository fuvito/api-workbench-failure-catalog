package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Event struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	CreatedAt    string `json:"created_at"`    // ISO 8601 date-time
	ScheduledFor string `json:"scheduled_for"` // ISO 8601 date
}

var events = []Event{
	{ID: 1, Name: "API Summit", CreatedAt: "2024-11-10T08:00:00Z", ScheduledFor: "2026-12-15"},
	{ID: 2, Name: "Platform Review", CreatedAt: "2024-11-11T08:00:00Z", ScheduledFor: "2027-01-08"},
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
