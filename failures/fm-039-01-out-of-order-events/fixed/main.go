package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"sync"
)

// An event is applied only if it is newer than the stored state and is an
// allowed transition. Stale events are acknowledged and ignored. An event that
// can't be applied yet gets 409, so the sender delivers it again later.

type event struct {
	OrderID  string `json:"order_id"`
	Status   string `json:"status"`
	Sequence int    `json:"sequence"`
}

type order struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Sequence int    `json:"sequence"`
}

// allowed maps a status to the statuses it may follow. "" means no order yet.
// "delivered" may follow "created" because "shipped" can arrive late.
var allowed = map[string][]string{
	"created":   {""},
	"shipped":   {"created"},
	"delivered": {"created", "shipped"},
	"cancelled": {"created"},
}

var (
	mu     sync.Mutex
	orders = map[string]order{}
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func receive(w http.ResponseWriter, r *http.Request) {
	var e event
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil || e.OrderID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid event"})
		return
	}
	mu.Lock()
	defer mu.Unlock()

	current, exists := orders[e.OrderID]
	if exists && e.Sequence <= current.Sequence {
		writeJSON(w, http.StatusOK, map[string]string{"result": "ignored"})
		return
	}
	if !slices.Contains(allowed[e.Status], current.Status) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "transition not allowed from current state"})
		return
	}
	orders[e.OrderID] = order{ID: e.OrderID, Status: e.Status, Sequence: e.Sequence}
	writeJSON(w, http.StatusOK, map[string]string{"result": "applied"})
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	o, ok := orders[r.PathValue("id")]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "order not found"})
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /events", receive)
	mux.HandleFunc("GET /orders/{id}", getOrder)

	http.ListenAndServe(":19065", mux)
}
