package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// The batch endpoint returns one result per item, in request order, tagged with
// the item's ref. A ref is paid at most once: sending a paid item again returns
// the original payment.

type item struct {
	Ref         string `json:"ref"`
	AmountCents int    `json:"amount_cents"`
}

type payment struct {
	ID          string `json:"id"`
	Ref         string `json:"ref"`
	AmountCents int    `json:"amount_cents"`
}

type itemError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"` // true if the same item may succeed when sent unchanged
}

type result struct {
	Ref       string     `json:"ref"`
	Status    string     `json:"status"`
	PaymentID string     `json:"payment_id,omitempty"`
	Error     *itemError `json:"error,omitempty"`
}

var (
	mu       sync.Mutex
	payments = []payment{}
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func rejected(ref, code, message string) result {
	return result{Ref: ref, Status: "failed", Error: &itemError{Code: code, Message: message}}
}

// process pays one item. The caller must hold mu.
func process(it item) result {
	if it.AmountCents <= 0 {
		return rejected(it.Ref, "invalid_amount", "amount_cents must be greater than 0")
	}
	for _, p := range payments {
		if p.Ref != it.Ref {
			continue
		}
		if p.AmountCents != it.AmountCents {
			return rejected(it.Ref, "ref_conflict", "ref was already paid with a different amount")
		}
		return result{Ref: it.Ref, Status: "succeeded", PaymentID: p.ID}
	}
	p := payment{ID: fmt.Sprintf("pmt-%d", len(payments)+1), Ref: it.Ref, AmountCents: it.AmountCents}
	payments = append(payments, p)
	return result{Ref: it.Ref, Status: "succeeded", PaymentID: p.ID}
}

func createBatch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Items []item `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Items) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "items is required"})
		return
	}
	// Without a ref an item can't be correlated or deduplicated, so check them all first.
	for _, it := range body.Items {
		if it.Ref == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "every item needs a ref"})
			return
		}
	}
	mu.Lock()
	defer mu.Unlock()

	results := make([]result, 0, len(body.Items))
	succeeded, failed := 0, 0
	for _, it := range body.Items {
		res := process(it)
		if res.Status == "succeeded" {
			succeeded++
		} else {
			failed++
		}
		results = append(results, res)
	}
	writeJSON(w, http.StatusOK, map[string]any{"succeeded": succeeded, "failed": failed, "results": results})
}

func listPayments(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{"count": len(payments), "payments": payments})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /payments/batch", createBatch)
	mux.HandleFunc("GET /payments", listPayments)

	http.ListenAndServe(":19053", mux)
}
