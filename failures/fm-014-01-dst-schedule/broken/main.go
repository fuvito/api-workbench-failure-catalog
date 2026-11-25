package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	_ "time/tzdata" // so the result doesn't depend on the machine's time zone files
)

// BUG: the schedule is turned into a fixed UTC instant and repeated every 24
// hours. After a daylight-saving change, "09:00 local" runs at 10:00 or 08:00.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
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
	first := time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute(), 0, 0, loc)

	runs := []string{}
	for i := range days {
		// BUG: a calendar day is not always 24 hours
		runs = append(runs, first.Add(time.Duration(i)*24*time.Hour).In(loc).Format(time.RFC3339))
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
