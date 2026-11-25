package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

// All CORS handling is in withCORS, which runs before authentication. Only the
// listed origin gets CORS headers, on preflights, successes and errors alike.

type order struct {
	ID   string `json:"id"`
	Note string `json:"note"`
}

// Synthetic bearer token for the one demo user.
const demoToken = "tok-alice"

// Origins allowed to call the API with credentials. Exact matches only.
var allowedOrigins = map[string]bool{
	"https://portal.example.test": true,
}

var (
	mu     sync.Mutex
	record = order{ID: "ord-1001", Note: "Leave at reception"}
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The headers depend on the Origin, so caches must key on it.
		w.Header().Add("Vary", "Origin")

		origin := r.Header.Get("Origin")
		allowed := allowedOrigins[origin]
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// A preflight carries no credentials, so answer it without authenticating.
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			if allowed {
				w.Header().Set("Access-Control-Allow-Methods", "GET, PUT")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if allowed {
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
		}
		next.ServeHTTP(w, r)
	})
}

// serve authenticates the caller, then runs the handler.
func serve(handle func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", r.Header.Get("X-Request-ID"))
		if r.Header.Get("Authorization") != "Bearer "+demoToken {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		handle(w, r)
	}
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	writeJSON(w, http.StatusOK, record)
}

func updateOrder(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Note string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Note == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "note is required"})
		return
	}
	mu.Lock()
	defer mu.Unlock()

	record.Note = body.Note
	writeJSON(w, http.StatusOK, record)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /orders/ord-1001", serve(getOrder))
	mux.HandleFunc("PUT /orders/ord-1001", serve(updateOrder))

	http.ListenAndServe(":19054", withCORS(mux))
}
