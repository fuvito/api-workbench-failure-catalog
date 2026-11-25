package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
)

// Both gateways are built from the same route and policy configuration.
//
// One process stands for both gateways: staging on :19074, production on :19075.

type gatewayConfig struct {
	ordersBackend  http.HandlerFunc
	reportsAuth    bool
	forwardTraceID bool
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// ---- backend services ----

func ordersV2(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"id": "ord-1001", "total": map[string]any{"amount": 49.5, "currency": "EUR"}})
}

// ---- gateway ----

func gateway(config gatewayConfig) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /v2/orders/ord-1001", config.ordersBackend)
	mux.HandleFunc("GET /v2/reports", func(w http.ResponseWriter, r *http.Request) {
		if config.reportsAuth && r.Header.Get("X-API-Key") != "demo-key-not-real" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]int{"count": 3})
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.forwardTraceID {
			if id := r.Header.Get("X-Request-ID"); id != "" {
				w.Header().Set("X-Request-ID", id)
			}
		}
		mux.ServeHTTP(w, r)
	})
}

func main() {
	release := gatewayConfig{ordersBackend: ordersV2, reportsAuth: true, forwardTraceID: true}

	stagingListener, err := net.Listen("tcp", ":19074")
	if err != nil {
		log.Fatal(err)
	}
	go func() { log.Fatal(http.Serve(stagingListener, gateway(release))) }()

	log.Fatal(http.ListenAndServe(":19075", gateway(release)))
}
