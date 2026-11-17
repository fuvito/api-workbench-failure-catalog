package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// BUG: POST /items/bulk-delete returns 204 but performs no deletion.
// The handler decodes the request body but never removes items from the store.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var items = map[int]Item{
	1: {ID: 1, Name: "Widget"},
	2: {ID: 2, Name: "Gadget"},
	3: {ID: 3, Name: "Doohickey"},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/items/bulk-delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			IDs []int `json:"ids"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		// BUG: IDs decoded but delete never executed
		_ = body.IDs
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/items/"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}
		item, ok := items[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusOK, item)
	})

	http.ListenAndServe(":19039", mux)
}
