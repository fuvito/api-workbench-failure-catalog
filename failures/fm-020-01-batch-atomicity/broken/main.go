package main

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

// BUG: batch insert is not atomic — valid items are created even when
// the batch contains an invalid one. The spec requires all-or-nothing.

var nextID int64 = 1

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type Item struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

var store []Item

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/items/batch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Items []struct{ Name string `json:"name"` } `json:"items"`
		}
		json.NewDecoder(r.Body).Decode(&body)

		// BUG: creates valid items and skips invalid instead of rejecting all
		created := 0
		for _, it := range body.Items {
			if it.Name == "" {
				continue // skip invalid, keep creating
			}
			store = append(store, Item{ID: atomic.AddInt64(&nextID, 1) - 1, Name: it.Name})
			created++
		}
		writeJSON(w, http.StatusCreated, map[string]int{"created": created})
	})

	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"items": store, "count": len(store)})
	})

	http.ListenAndServe(":19020", mux)
}
