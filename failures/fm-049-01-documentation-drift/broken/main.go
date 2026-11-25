package main

import (
	"encoding/json"
	"net/http"
)

// BUG: the API changed but its documentation didn't. The documented examples
// use the old names, so copying them no longer works.

type ticket struct {
	ID       string `json:"id"`
	Subject  string `json:"subject"`
	Priority string `json:"priority"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func authorized(r *http.Request) bool { return r.Header.Get("Api-Key") == "demo-key-not-real" }

func create(w http.ResponseWriter, r *http.Request) {
	if !authorized(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	// BUG: "subject" was renamed to "title" in the code only
	var body struct {
		Title    string `json:"title"`
		Priority string `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	writeJSON(w, http.StatusCreated, ticket{ID: "tkt-1001", Subject: body.Title, Priority: body.Priority})
}

func list(w http.ResponseWriter, r *http.Request) {
	if !authorized(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	// BUG: the documented filter "priority=high" is no longer accepted
	if r.URL.Query().Has("priority") {
		writeError(w, http.StatusBadRequest, "unknown parameter: priority")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tickets": []ticket{{ID: "tkt-1001", Subject: "Printer is offline", Priority: "high"}}})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /tickets", create)
	mux.HandleFunc("GET /tickets", list)

	http.ListenAndServe(":19081", mux)
}
