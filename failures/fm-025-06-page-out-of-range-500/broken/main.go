package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// BUG: when page is beyond the last page, the server returns 500 instead of
// 200 with an empty list. No bounds check is applied before slicing.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var allItems = []Item{
	{ID: 1, Name: "Alpha"},
	{ID: 2, Name: "Beta"},
	{ID: 3, Name: "Gamma"},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		perPage := 10

		start := (page - 1) * perPage
		// BUG: no bounds check — start beyond len(allItems) causes a 500
		if start > len(allItems) {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		end := start + perPage
		if end > len(allItems) {
			end = len(allItems)
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"items": allItems[start:end],
			"count": len(allItems[start:end]),
			"page":  page,
		})
	})

	http.ListenAndServe(":19050", mux)
}
