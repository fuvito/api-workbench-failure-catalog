package main

import (
	"encoding/json"
	"net/http"
)

// BUG: the /settings endpoint accepts any HTTP method.
// DELETE /settings should return 405 Method Not Allowed,
// but the broken server processes it and returns 200.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

var settings = map[string]any{
	"theme":    "dark",
	"language": "en",
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// BUG: no method guard — responds 200 to any method including DELETE
	mux.HandleFunc("/settings", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, settings)
	})

	http.ListenAndServe(":19014", mux)
}
