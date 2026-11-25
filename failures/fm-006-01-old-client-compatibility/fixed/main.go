package main

import (
	"encoding/json"
	"net/http"
)

// New behaviour is opt-in with "Api-Version: 2". Without it, the response keeps
// the contract that released apps were built against: a status they know and
// tracking_code where they read it.

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Vary", "Api-Version")
	shipment := map[string]string{"tracking_code": "TRK-1", "pickup_point": "Locker 7"}

	if r.Header.Get("Api-Version") == "2" {
		writeJSON(w, map[string]any{"id": "ord-1001", "status": "awaiting_pickup", "shipment": shipment})
		return
	}
	writeJSON(w, map[string]any{
		"id": "ord-1001",
		// To an old app, an order waiting for pickup has been shipped.
		"status":        "shipped",
		"tracking_code": shipment["tracking_code"],
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /orders/ord-1001", getOrder)

	http.ListenAndServe(":19079", mux)
}
