package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

// BUG: the CORS handling works for the simplest request only. The preflight
// does not allow Authorization, X-Request-ID or PUT, error responses carry no
// CORS headers, X-Request-ID is not exposed, and any origin is allowed to make
// credentialed requests.

type order struct {
	ID   string `json:"id"`
	Note string `json:"note"`
}

// Synthetic bearer token for the one demo user.
const demoToken = "tok-alice"

var (
	mu     sync.Mutex
	record = order{ID: "ord-1001", Note: "Leave at reception"}
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func allowOrigin(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		// BUG: echoes back whatever origin asked, and allows credentials
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		// BUG: no Vary: Origin, so a cache can serve one origin's headers to another
	}
}

func preflight(w http.ResponseWriter, r *http.Request) {
	allowOrigin(w, r)
	// BUG: PUT is missing
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	// BUG: Authorization and X-Request-ID are missing
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusNoContent)
}

// serve authenticates the caller, then runs the handler.
func serve(handle func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", r.Header.Get("X-Request-ID"))
		if r.Header.Get("Authorization") != "Bearer "+demoToken {
			// BUG: returns before the CORS headers are added
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		// BUG: no Access-Control-Expose-Headers, so a page cannot read X-Request-ID
		allowOrigin(w, r)
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
	mux.HandleFunc("OPTIONS /orders/ord-1001", preflight)
	mux.HandleFunc("GET /orders/ord-1001", serve(getOrder))
	mux.HandleFunc("PUT /orders/ord-1001", serve(updateOrder))

	http.ListenAndServe(":19054", mux)
}
