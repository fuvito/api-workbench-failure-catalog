package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// A new key leaves the old one working, marked "retiring", until it is revoked
// explicitly. A secret is shown only once, when the key is created.

type key struct {
	ID     string `json:"id"`
	Secret string `json:"secret,omitempty"`
	Status string `json:"status"`
}

// All keys are synthetic.
var (
	mu   sync.Mutex
	keys = []*key{{ID: "key-1", Secret: "demo-key-1-not-real", Status: "active"}}
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func usable(r *http.Request) bool {
	secret := r.Header.Get("X-API-Key")
	mu.Lock()
	defer mu.Unlock()
	for _, k := range keys {
		if k.Secret == secret && k.Status != "revoked" {
			return true
		}
	}
	return false
}

func admin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") != "tok-admin" {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r)
	}
}

func rotate(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	for _, k := range keys {
		if k.Status == "active" {
			k.Status = "retiring"
		}
	}
	n := len(keys) + 1
	created := &key{ID: fmt.Sprintf("key-%d", n), Secret: fmt.Sprintf("demo-key-%d-not-real", n), Status: "active"}
	keys = append(keys, created)
	writeJSON(w, http.StatusCreated, created)
}

func list(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	listed := make([]key, len(keys))
	for i, k := range keys {
		listed[i] = key{ID: k.ID, Status: k.Status}
	}
	writeJSON(w, http.StatusOK, map[string]any{"keys": listed})
}

func revoke(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	for _, k := range keys {
		if k.ID == r.PathValue("id") {
			k.Status = "revoked"
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	writeError(w, http.StatusNotFound, "key not found")
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /reports", func(w http.ResponseWriter, r *http.Request) {
		if !usable(r) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeJSON(w, http.StatusOK, map[string]int{"count": 3})
	})
	mux.HandleFunc("POST /admin/keys", admin(rotate))
	mux.HandleFunc("GET /admin/keys", admin(list))
	mux.HandleFunc("DELETE /admin/keys/{id}", admin(revoke))

	http.ListenAndServe(":19069", mux)
}
