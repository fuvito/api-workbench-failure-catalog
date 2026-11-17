package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// BUG: each handler returns a different error shape on 404.
// The spec declares a single Error schema: {"error": string}.
// - /users/:id returns {"message": "user not found"}
// - /orders/:id returns {"code": 404, "msg": "not found"}
// - /products/:id returns {"error": "not found"} — this one accidentally matches

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/users/"))
		if err != nil || id != 1 {
			// BUG: uses "message" instead of "error"
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "user not found"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"id": 1, "name": "Alice", "email": "alice@example.com"})
	})

	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/orders/"))
		if err != nil || id != 1 {
			// BUG: uses numeric "code" + "msg" instead of string "error"
			writeJSON(w, http.StatusNotFound, map[string]any{"code": 404, "msg": "not found"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"id": 1, "user_id": 1, "total": 49.99})
	})

	mux.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/products/"))
		if err != nil || id != 1 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"id": 1, "name": "Widget", "price": 9.99})
	})

	http.ListenAndServe(":19006", mux)
}
