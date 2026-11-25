package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	_ "time/tzdata" // so the result doesn't depend on the machine's time zone files
)

// Each run is worked out from its own calendar date and the local wall-clock
// time. Policy: a time that occurs twice runs once, at the first occurrence; a
// time that doesn't exist runs when the clocks have moved on.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// resolve returns the instant for a wall-clock time (given as if it were UTC)
// in loc, applying the policy above.
func resolve(wall time.Time, loc *time.Location) time.Time {
	// The offsets in force a day before and a day after cover any transition.
	_, before := wall.Add(-24 * time.Hour).In(loc).Zone()
	_, after := wall.Add(24 * time.Hour).In(loc).Zone()

	var found time.Time
	for _, offset := range []int{before, after} {
		candidate := wall.Add(-time.Duration(offset) * time.Second).In(loc)
		sameWall := candidate.Format("2006-01-02T15:04") == wall.Format("2006-01-02T15:04")
		if sameWall && (found.IsZero() || candidate.Before(found)) {
			found = candidate
		}
	}
	if found.IsZero() {
		// Skipped by the clock change: use the old offset, which lands after it.
		found = wall.Add(-time.Duration(before) * time.Second).In(loc)
	}
	return found
}

func preview(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	loc, errZone := time.LoadLocation(q.Get("timezone"))
	start, errStart := time.Parse("2006-01-02T15:04", q.Get("from")+"T"+q.Get("time"))
	days, errDays := strconv.Atoi(q.Get("days"))
	if errZone != nil || errStart != nil || errDays != nil || days < 1 || days > 7 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid schedule"})
		return
	}
	runs := []string{}
	for i := range days {
		// AddDate moves by calendar days. start is in UTC, which has no clock changes.
		runs = append(runs, resolve(start.AddDate(0, 0, i), loc).Format(time.RFC3339))
	}
	writeJSON(w, http.StatusOK, map[string]any{"timezone": loc.String(), "runs": runs})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /schedule/preview", preview)

	http.ListenAndServe(":19080", mux)
}
