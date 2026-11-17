package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// BUG: two HTTP method semantics violations:
//  1. POST /articles returns 200 OK — should be 201 Created.
//  2. DELETE /articles/{id} returns 200 with a JSON body — should be 204 No Content.
//
// Both status codes are absent from openapi.yaml (which declares 201 and 204 only),
// so the specFile: binding catches them as undeclared responses.

type Article struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

type ArticleInput struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

var (
	nextID   = 2
	articles = map[int]Article{
		1: {ID: 1, Title: "Hello World", Author: "Alice"},
	}
)

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

	mux.HandleFunc("POST /articles", func(w http.ResponseWriter, r *http.Request) {
		var input ArticleInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid body", Code: "INVALID_BODY"})
			return
		}
		id := nextID
		nextID++
		article := Article{ID: id, Title: input.Title, Author: input.Author}
		articles[id] = article
		writeJSON(w, http.StatusOK, article) // BUG: should be http.StatusCreated (201)
	})

	mux.HandleFunc("GET /articles/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid id", Code: "INVALID_ID"})
			return
		}
		article, ok := articles[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not found", Code: "ARTICLE_NOT_FOUND"})
			return
		}
		writeJSON(w, http.StatusOK, article)
	})

	mux.HandleFunc("DELETE /articles/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid id", Code: "INVALID_ID"})
			return
		}
		if _, ok := articles[id]; !ok {
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not found", Code: "ARTICLE_NOT_FOUND"})
			return
		}
		delete(articles, id)
		writeJSON(w, http.StatusOK, map[string]bool{"deleted": true}) // BUG: should be 204 No Content with no body
	})

	log.Fatal(http.ListenAndServe(":19034", mux))
}
