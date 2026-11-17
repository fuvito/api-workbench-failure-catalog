package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// BUG: PUT /items/{id} ignores the version field in the request body and always
// overwrites. Two clients that both read version=1 and then write will silently
// clobber each other — the second write wins and the first is lost.
//
// The spec requires optimistic locking: if the submitted version does not match
// the current version, the server must return 409 Conflict so the client can
// re-read and merge before retrying.

type Item struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Value   int    `json:"value"`
	Version int    `json:"version"`
}

type ItemUpdate struct {
	Value   int `json:"value"`
	Version int `json:"version"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

var items = map[int]*Item{
	1: {ID: 1, Name: "Widget", Value: 10, Version: 1},
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid id", Code: "INVALID_ID"})
			return
		}
		item, ok := items[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not found", Code: "ITEM_NOT_FOUND"})
			return
		}
		writeJSON(w, http.StatusOK, item)
	})

	mux.HandleFunc("PUT /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid id", Code: "INVALID_ID"})
			return
		}
		item, ok := items[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not found", Code: "ITEM_NOT_FOUND"})
			return
		}
		var update ItemUpdate
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid body", Code: "INVALID_BODY"})
			return
		}
		// BUG: version is never checked — stale writes silently overwrite current data
		item.Value = update.Value
		item.Version++
		writeJSON(w, http.StatusOK, item)
	})

	log.Fatal(http.ListenAndServe(":19027", mux))
}
