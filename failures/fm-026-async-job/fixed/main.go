package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
)

var nextID int64 = 1

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type Job struct {
	ID     int64  `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

var jobs = map[int64]*Job{}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var body struct{ Type string `json:"type"` }
		json.NewDecoder(r.Body).Decode(&body)
		id := atomic.AddInt64(&nextID, 1) - 1
		// Job completes synchronously in this simplified server
		j := &Job{ID: id, Type: body.Type, Status: "completed"}
		jobs[id] = j
		writeJSON(w, http.StatusCreated, j)
	})

	mux.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/jobs/"), 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}
		j, ok := jobs[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
			return
		}
		writeJSON(w, http.StatusOK, j)
	})

	http.ListenAndServe(":19026", mux)
}
