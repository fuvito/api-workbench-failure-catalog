package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

// BUG: internal details reach callers through diagnostic headers, error
// messages, an extra "detail" field and an unauthenticated debug endpoint.

// Every value here is fake.
const (
	demoToken  = "tok-alice"
	dbHost     = "db01.internal.example.test"
	dbDSN      = "postgres://reports_app:demo-db-password@" + dbHost + ":5432/reports"
	signingKey = "demo-signing-key-not-real"
	reportSQL  = "SELECT id, title FROM reports WHERE id = $1"
)

type report struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// queryReport stands in for a database call. rpt-2002 always fails.
func queryReport(id string) (report, bool, error) {
	switch id {
	case "rpt-1001":
		return report{ID: id, Title: "Quarterly summary"}, true, nil
	case "rpt-2002":
		return report{}, false, errors.New("query failed: connection to " + dbDSN + " reset")
	}
	return report{}, false, nil
}

func getReport(w http.ResponseWriter, r *http.Request) {
	// BUG: diagnostic headers left on in every response
	w.Header().Set("X-Backend-Host", dbHost)
	w.Header().Set("X-Debug-Query", reportSQL)

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token != demoToken {
		// BUG: echoes the credential back
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unknown token: " + token})
		return
	}

	if from := r.URL.Query().Get("from"); from != "" {
		if _, err := time.Parse(time.DateOnly, from); err != nil {
			// BUG: the bad value reaches the database, and its error reaches the caller
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "invalid input syntax for type date: \"" + from + "\" (" + dbDSN + ")",
			})
			return
		}
	}

	rep, found, err := queryReport(r.PathValue("id"))
	if err != nil {
		// BUG: the message is safe, but the cause is attached beside it
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":  "internal error",
			"detail": err.Error(),
		})
		return
	}
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "report not found"})
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
	// BUG: a debug endpoint with no authentication
	mux.HandleFunc("GET /debug/config", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"database_url": dbDSN, "signing_key": signingKey})
	})

	http.ListenAndServe(":19058", mux)
}
