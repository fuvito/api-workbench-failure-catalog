package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// Only text meant for people is translated. Codes, enum values, numbers and
// dates are the same in every language. Unsupported languages fall back to
// English.

type texts struct{ inProgress, notFound string }

var translations = map[string]texts{
	"en": {inProgress: "In progress", notFound: "Order not found"},
	"fr": {inProgress: "En cours", notFound: "Commande introuvable"},
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// language picks the supported language with the highest quality value in
// Accept-Language, matching on the primary subtag (fr-CH matches fr).
func language(header string) string {
	best, bestQ := "en", 0.0
	for _, entry := range strings.Split(header, ",") {
		tag, params, _ := strings.Cut(strings.TrimSpace(entry), ";")
		q := 1.0
		if value, ok := strings.CutPrefix(strings.TrimSpace(params), "q="); ok {
			parsed, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}
			q = parsed
		}
		primary, _, _ := strings.Cut(strings.ToLower(tag), "-")
		if _, ok := translations[primary]; ok && q > bestQ {
			best, bestQ = primary, q
		}
	}
	return best
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	lang := language(r.Header.Get("Accept-Language"))
	w.Header().Set("Content-Language", lang)
	w.Header().Set("Vary", "Accept-Language")

	if r.PathValue("id") != "ord-1001" {
		writeJSON(w, http.StatusNotFound, map[string]string{"code": "ORDER_NOT_FOUND", "message": translations[lang].notFound})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id": "ord-1001", "status": "in_progress", "status_label": translations[lang].inProgress,
		"total": 1234.5, "created_at": "2026-03-04T09:30:00Z",
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /orders/{id}", getOrder)

	http.ListenAndServe(":19061", mux)
}
