package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
)

// Search text is compared as data and never parsed. The owner condition comes
// from the caller's identity, and sort accepts only listed columns.

type product struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

var table = []product{
	{Name: "Desk lamp", Owner: "alice"},
	{Name: "O'Brien chair", Owner: "alice"},
	{Name: "Server rack", Owner: "bob"},
}

// sortable lists the columns a caller may sort by.
var sortable = map[string]func(a, b product) int{
	"name": func(a, b product) int { return strings.Compare(a.Name, b.Name) },
}

// search is the equivalent of a parameterized query:
// WHERE owner = ? AND name LIKE ?
func search(owner, text string) []product {
	out := []product{}
	for _, p := range table {
		if p.Owner == owner && strings.Contains(p.Name, text) {
			out = append(out, p)
		}
	}
	return out
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func listProducts(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") != "Bearer tok-alice" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	var compare func(a, b product) int
	if key := r.URL.Query().Get("sort"); key != "" {
		var ok bool
		if compare, ok = sortable[key]; !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported sort field"})
			return
		}
	}
	items := search("alice", r.URL.Query().Get("search"))
	if compare != nil {
		slices.SortStableFunc(items, compare)
	}
	writeJSON(w, http.StatusOK, map[string]any{"count": len(items), "items": items})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /products", listProducts)

	http.ListenAndServe(":19064", mux)
}
