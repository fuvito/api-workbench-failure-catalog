package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
)

// BUG: user input is pasted into a filter expression, which is then parsed.
// Input containing " OR " changes the query. The sort parameter is also used
// unchecked, so callers can sort by hidden columns.
//
// The tiny query language below stands in for SQL. Nothing here can run
// commands or touch anything outside this process.

type row struct {
	fields map[string]string
}

var table = []row{
	{map[string]string{"owner": "alice", "name": "Desk lamp", "cost": "12"}},
	{map[string]string{"owner": "alice", "name": "O'Brien chair", "cost": "40"}},
	{map[string]string{"owner": "bob", "name": "Server rack", "cost": "900"}},
}

// query evaluates "a=x AND b~y OR c=z": OR binds weakest, = is equality and ~
// is "contains".
func query(expression string) []row {
	var out []row
	for _, r := range table {
		for _, branch := range strings.Split(expression, " OR ") {
			all := true
			for _, condition := range strings.Split(branch, " AND ") {
				if field, value, ok := strings.Cut(condition, "="); ok {
					all = all && r.fields[field] == value
				} else if field, value, ok := strings.Cut(condition, "~"); ok {
					all = all && strings.Contains(r.fields[field], value)
				} else {
					all = false
				}
			}
			if all {
				out = append(out, r)
				break
			}
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
	// BUG: the search text becomes part of the expression
	rows := query("owner=alice AND name~" + r.URL.Query().Get("search"))

	// BUG: any column name is accepted, including ones never returned
	if key := r.URL.Query().Get("sort"); key != "" {
		slices.SortStableFunc(rows, func(a, b row) int { return strings.Compare(a.fields[key], b.fields[key]) })
	}

	items := []map[string]string{}
	for _, found := range rows {
		items = append(items, map[string]string{"name": found.fields["name"], "owner": found.fields["owner"]})
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
