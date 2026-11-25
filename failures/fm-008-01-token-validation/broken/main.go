package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
)

// BUG: only the signature is checked, and not even that for alg "none". Expiry,
// issuer, audience and scope are ignored.

// Synthetic signing secret.
var secret = []byte("demo-jwt-secret-not-real")

type claims struct {
	Issuer   string `json:"iss"`
	Audience string `json:"aud"`
	Subject  string `json:"sub"`
	Scope    string `json:"scope"`
	Expires  int64  `json:"exp"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func decode(segment string, into any) bool {
	raw, err := base64.RawURLEncoding.DecodeString(segment)
	return err == nil && json.Unmarshal(raw, into) == nil
}

func verify(token string) (claims, bool) {
	var c claims
	parts := strings.Split(token, ".")
	var header struct {
		Alg string `json:"alg"`
	}
	if len(parts) != 3 || !decode(parts[0], &header) || !decode(parts[1], &c) {
		return c, false
	}
	// BUG: trusts the token to say it needs no signature
	if header.Alg == "none" {
		return c, true
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	got, err := base64.RawURLEncoding.DecodeString(parts[2])
	// BUG: exp, iss and aud are never checked
	return c, err == nil && hmac.Equal(got, mac.Sum(nil))
}

// authorize ignores the required scope.
func authorize(scope string, next func(w http.ResponseWriter, c claims)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		c, ok := verify(token)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_token"})
			return
		}
		// BUG: authenticated is treated as authorized
		next(w, c)
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /payments", authorize("payments:read", func(w http.ResponseWriter, c claims) {
		writeJSON(w, http.StatusOK, map[string]any{"subject": c.Subject, "count": 2})
	}))
	mux.HandleFunc("POST /refunds", authorize("payments:write", func(w http.ResponseWriter, c claims) {
		writeJSON(w, http.StatusCreated, map[string]string{"status": "refunded"})
	}))

	http.ListenAndServe(":19078", mux)
}
