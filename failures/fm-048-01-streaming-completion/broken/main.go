package main

import (
	"fmt"
	"net/http"
	"strconv"
)

// BUG: when the data source fails partway through a streamed export, the server
// closes the JSON neatly and says it is complete. It also ignores "after", so an
// export can't be resumed.

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
	_ = after // BUG: never used
	failing := r.Header.Get("X-Test-Fault") == "source-fails-after-2"

	w.Header().Set("Content-Type", "application/json")
	flusher, _ := w.(http.Flusher)
	fmt.Fprint(w, `{"records":[`)
	last := 0
	for id := 1; id <= total; id++ {
		if !readRecord(id, failing) {
			break // BUG: the error is swallowed
		}
		if last > 0 {
			fmt.Fprint(w, ",")
		}
		fmt.Fprintf(w, `{"id":%d}`, id)
		last = id
		if flusher != nil {
			flusher.Flush()
		}
	}
	// BUG: always claims to be complete
	fmt.Fprintf(w, `],"complete":true,"next_after":%d}`, last)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /exports/records", export)

	http.ListenAndServe(":19083", mux)
}
