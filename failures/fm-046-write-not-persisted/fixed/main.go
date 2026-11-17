package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

var (
	mu     sync.Mutex
	config = map[string]string{
		"theme":    "light",
		"language": "en",
	}
)

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

	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, config)

		case http.MethodPut:
			var update map[string]string
			json.NewDecoder(r.Body).Decode(&update)
			for k, v := range update {
				config[k] = v
			}
			writeJSON(w, http.StatusOK, config)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.ListenAndServe(":19046", mux)
}
