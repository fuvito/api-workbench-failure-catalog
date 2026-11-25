package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// BUG: Accept-Language changes values that programs read, not just text for
// people: the status, the amount, the date and the error code.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func french(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("Accept-Language"), "fr")
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("id") != "ord-1001" {
		code := "ORDER_NOT_FOUND"
		message := "Order not found"
		if french(r) {
			// BUG: the code is translated along with the message
			code, message = "COMMANDE_INTROUVABLE", "Commande introuvable"
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"code": code, "message": message})
		return
	}
	order := map[string]any{
		"id": "ord-1001", "status": "in_progress", "status_label": "In progress",
		"total": 1234.5, "created_at": "2026-03-04T09:30:00Z",
	}
	if french(r) {
		// BUG: status, number and date are formatted for a person
		order["status"] = "en_cours"
		order["status_label"] = "En cours"
		order["total"] = "1 234,50"
		order["created_at"] = "04/03/2026 09:30"
	}
	writeJSON(w, http.StatusOK, order)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /orders/{id}", getOrder)

	http.ListenAndServe(":19061", mux)
}
