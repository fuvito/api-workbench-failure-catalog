package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

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
		totalPages := (len(allItems) + perPage - 1) / perPage

		start := (page - 1) * perPage
		if start >= len(allItems) {
			// Page beyond data — return empty result
			writeJSON(w, http.StatusOK, map[string]any{
				"items":       []Item{},
				"count":       0,
				"page":        page,
				"total_pages": totalPages,
			})
			return
		}
		end := start + perPage
		if end > len(allItems) {
			end = len(allItems)
		}
		pageItems := allItems[start:end]
		writeJSON(w, http.StatusOK, map[string]any{
			"items":       pageItems,
			"count":       len(pageItems),
			"page":        page,
			"total_pages": totalPages,
		})
	})

	http.ListenAndServe(":19031", mux)
}
