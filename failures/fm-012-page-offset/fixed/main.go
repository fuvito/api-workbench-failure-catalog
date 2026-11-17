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
	{ID: 4, Name: "Delta"},
	{ID: 5, Name: "Epsilon"},
	{ID: 6, Name: "Zeta"},
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
		perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
		if perPage < 1 {
			perPage = 3
		}

		start := (page - 1) * perPage
		if start > len(allItems) {
			start = len(allItems)
		}
		end := start + perPage
		if end > len(allItems) {
			end = len(allItems)
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"items":    allItems[start:end],
			"count":    len(allItems[start:end]),
			"page":     page,
			"per_page": perPage,
		})
	})

	http.ListenAndServe(":19012", mux)
}
