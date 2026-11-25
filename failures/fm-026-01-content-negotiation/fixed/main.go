package main

import (
	"encoding/json"
	"mime"
	"net/http"
	"strconv"
	"strings"
)

// The response type is chosen from Accept, with quality values and wildcards. A
// request nothing can satisfy gets 406. The request Content-Type is parsed, so
// parameters don't matter, and other media types get 415.

const (
	v1 = "application/json"
	v2 = "application/vnd.example.order-v2+json"
)

// offered is in order of preference: it breaks ties, so */* gets v1.
var offered = []string{v1, v2}

func write(w http.ResponseWriter, status int, contentType string, v any) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// quality returns how much the Accept header wants mediaType. The most specific
// matching range decides.
func quality(accept, mediaType string) float64 {
	mainType, _, _ := strings.Cut(mediaType, "/")
	best, bestRank := 0.0, 0
	for _, entry := range strings.Split(accept, ",") {
		name, params, err := mime.ParseMediaType(strings.TrimSpace(entry))
		if err != nil {
			continue
		}
		rank := 0
		switch name {
		case mediaType:
			rank = 3
		case mainType + "/*":
			rank = 2
		case "*/*":
			rank = 1
		}
		if rank <= bestRank {
			continue
		}
		q := 1.0
		if value, ok := params["q"]; ok {
			if q, err = strconv.ParseFloat(value, 64); err != nil {
				continue
			}
		}
		best, bestRank = q, rank
	}
	return best
}

func negotiate(accept string) (string, bool) {
	if accept == "" {
		return v1, true
	}
	chosen, chosenQ := "", 0.0
	for _, mediaType := range offered {
		if q := quality(accept, mediaType); q > chosenQ {
			chosen, chosenQ = mediaType, q
		}
	}
	return chosen, chosenQ > 0
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Vary", "Accept")
	mediaType, ok := negotiate(r.Header.Get("Accept"))
	if !ok {
		write(w, http.StatusNotAcceptable, v1, map[string]string{"error": "not acceptable"})
		return
	}
	if mediaType == v2 {
		write(w, http.StatusOK, v2, map[string]any{"id": "ord-1001", "total": map[string]any{"amount": 49.5, "currency": "EUR"}})
		return
	}
	write(w, http.StatusOK, v1, map[string]any{"id": "ord-1001", "total": 49.5})
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != v1 {
		write(w, http.StatusUnsupportedMediaType, v1, map[string]string{"error": "unsupported media type"})
		return
	}
	var body struct {
		Total float64 `json:"total"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Total <= 0 {
		write(w, http.StatusBadRequest, v1, map[string]string{"error": "total is required"})
		return
	}
	write(w, http.StatusCreated, v1, map[string]any{"id": "ord-1002", "total": body.Total})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /orders/ord-1001", getOrder)
	mux.HandleFunc("POST /orders", createOrder)

	http.ListenAndServe(":19062", mux)
}
