package main

import (
	"encoding/json"
	"net/http"
)

// BUG: version 2 behaviour is sent to every client. Older apps, which send no
// Api-Version header, get a status they don't know and lose tracking_code.

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	// BUG: Api-Version is never read
	writeJSON(w, map[string]any{
		"id":       "ord-1001",
		"status":   "awaiting_pickup",
		"shipment": map[string]string{"tracking_code": "TRK-1", "pickup_point": "Locker 7"},
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
