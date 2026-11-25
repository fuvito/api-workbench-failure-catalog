package main

import (
	"fmt"
	"net/http"
	"strconv"
)

// The end of the stream says whether it is complete. If the data source fails
// partway, complete is false and next_after tells the client where to resume.

const total = 5

// readRecord stands in for reading one row. The test fault makes it fail after
// the second record.
func readRecord(id int, failing bool) bool { return !(failing && id > 2) }

func export(w http.ResponseWriter, r *http.Request) {
	after := 0
	if value := r.URL.Query().Get("after"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 || parsed > total {
			http.Error(w, `{"error":"invalid after"}`, http.StatusBadRequest)
			return
		}
		after = parsed
	}
	failing := r.Header.Get("X-Test-Fault") == "source-fails-after-2"

	w.Header().Set("Content-Type", "application/json")
	flusher, _ := w.(http.Flusher)
	fmt.Fprint(w, `{"records":[`)
	last, complete := after, true
	for id := after + 1; id <= total; id++ {
		if !readRecord(id, failing) {
			complete = false
			break
		}
		if last > after {
			fmt.Fprint(w, ",")
		}
		fmt.Fprintf(w, `{"id":%d}`, id)
		last = id
		if flusher != nil {
			flusher.Flush()
		}
	}
	fmt.Fprintf(w, `],"complete":%t,"next_after":%d}`, complete, last)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /exports/records", export)

	http.ListenAndServe(":19083", mux)
}
