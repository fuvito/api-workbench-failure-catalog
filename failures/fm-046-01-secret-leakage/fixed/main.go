package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

// Callers get a fixed message for each kind of error. The cause of a server
// error stays on the server, and even the log leaves out the connection string.

// Every value here is fake.
const (
	demoToken = "tok-alice"
)

var errDatabase = errors.New("database connection reset")

type report struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError takes only fixed strings, never an error value or request data.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// queryReport stands in for a database call. rpt-2002 always fails.
func queryReport(id string) (report, bool, error) {
	switch id {
	case "rpt-1001":
		return report{ID: id, Title: "Quarterly summary"}, true, nil
	case "rpt-2002":
		return report{}, false, errDatabase
	}
	return report{}, false, nil
}

func getReport(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") != "Bearer "+demoToken {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Validate before the value gets anywhere near the database.
	if from := r.URL.Query().Get("from"); from != "" {
		if _, err := time.Parse(time.DateOnly, from); err != nil {
			writeError(w, http.StatusBadRequest, "from must be a date (YYYY-MM-DD)")
			return
		}
	}

	rep, found, err := queryReport(r.PathValue("id"))
	if err != nil {
		log.Printf("get report: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "report not found")
		return
	}
	writeJSON(w, http.StatusOK, rep)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /reports/{id}", getReport)

	http.ListenAndServe(":19058", mux)
}
