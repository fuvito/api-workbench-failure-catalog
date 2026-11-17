package main

import (
	"encoding/json"
	"net/http"
)

// BUG: the tags list contains an integer (42) alongside string values.
// A type-assertion bug in the data layer accidentally serialises a numeric ID
// as an array element instead of converting it to a string.

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

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		// BUG: 42 is an integer, not a string
		writeJSON(w, http.StatusOK, map[string]any{
			"items": []any{"go", "api", 42, "rest"},
			"count": 4,
		})
	})

	http.ListenAndServe(":19015", mux)
}
