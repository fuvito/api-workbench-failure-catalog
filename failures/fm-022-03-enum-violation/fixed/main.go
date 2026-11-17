package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type Order struct {
	ID       int    `json:"id"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
}

var orders = map[int]Order{
	1: {ID: 1, Status: "pending", Priority: "high"},
	2: {ID: 2, Status: "processing", Priority: "medium"},
	3: {ID: 3, Status: "completed", Priority: "high"},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/orders/"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}
		o, ok := orders[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "order not found"})
			return
		}
		writeJSON(w, http.StatusOK, o)
	})

	http.ListenAndServe(":19030", mux)
}
