package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
)

// Every instance in every environment is started from the same versioned
// release config, and reports that version so drift can be detected.
//
// One process stands for both environments: staging on :19072 and production
// on :19073.

type config struct {
	Version string `json:"config_version"`
	NewTax  bool   `json:"new_tax_rule"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// environment serves requests from its instances in turn.
func environment(instances ...config) http.Handler {
	var requests atomic.Int64
	pick := func() config { return instances[int(requests.Add(1)-1)%len(instances)] }

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /quote", func(w http.ResponseWriter, r *http.Request) {
		cents, err := strconv.Atoi(r.URL.Query().Get("subtotal_cents"))
		if err != nil || cents <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid subtotal"})
			return
		}
		rate := 8
		if pick().NewTax {
			rate = 10
		}
		writeJSON(w, http.StatusOK, map[string]int{"total_cents": cents + cents*rate/100})
	})
	mux.HandleFunc("GET /config", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, pick())
	})
	return mux
}

func main() {
	release := config{Version: "2026-11-20", NewTax: true}
	staging, err := net.Listen("tcp", ":19072")
	if err != nil {
		log.Fatal(err)
	}
	go func() { log.Fatal(http.Serve(staging, environment(release))) }()

	log.Fatal(http.ListenAndServe(":19073", environment(release, release, release)))
}
