package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

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
		// Optimistic locking: reject stale writes so the client must re-read and merge.
		if update.Version != item.Version {
			writeJSON(w, http.StatusConflict, ErrorResponse{Error: "version conflict", Code: "VERSION_CONFLICT"})
			return
		}
		item.Value = update.Value
		item.Version++
		writeJSON(w, http.StatusOK, item)
	})

	log.Fatal(http.ListenAndServe(":19027", mux))
}
