package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Payment struct {
	ID       int     `json:"id"`
	OrderID  string  `json:"order_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Status   string  `json:"status"`
}

type PaymentInput struct {
	OrderID  string  `json:"order_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type PaymentList struct {
	Items []Payment `json:"items"`
	Count int       `json:"count"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

var (
	nextID   = 1
	payments []Payment
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("POST /payments", func(w http.ResponseWriter, r *http.Request) {
		var input PaymentInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid body", Code: "INVALID_BODY"})
			return
		}
		if input.OrderID == "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "order_id required", Code: "MISSING_ORDER_ID"})
			return
		}
		// Idempotency check: return the existing payment if this order_id was already processed.
		for _, p := range payments {
			if p.OrderID == input.OrderID {
				writeJSON(w, http.StatusOK, p)
				return
			}
		}
		payment := Payment{
			ID:       nextID,
			OrderID:  input.OrderID,
			Amount:   input.Amount,
			Currency: input.Currency,
			Status:   "completed",
		}
		nextID++
		payments = append(payments, payment)
		writeJSON(w, http.StatusCreated, payment)
	})

	mux.HandleFunc("GET /payments", func(w http.ResponseWriter, r *http.Request) {
		if payments == nil {
			payments = []Payment{}
		}
		writeJSON(w, http.StatusOK, PaymentList{Items: payments, Count: len(payments)})
	})

	log.Fatal(http.ListenAndServe(":19018", mux))
}
