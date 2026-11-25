package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"
)

// A token must be signed with HS256 by our key, unexpired, from our issuer and
// for this API (401 otherwise), and carry the scope the route needs (403).

const (
	issuer   = "https://auth.example.test"
	audience = "payments-api"
)

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
	// The server decides the algorithm, not the token.
	if header.Alg != "HS256" {
		return c, false
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	got, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(got, mac.Sum(nil)) {
		return c, false
	}
	return c, c.Issuer == issuer && c.Audience == audience && time.Now().Unix() < c.Expires
}

// authorize requires a valid token that has the scope.
func authorize(scope string, next func(w http.ResponseWriter, c claims)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		c, ok := verify(token)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_token"})
			return
		}
		if !slices.Contains(strings.Fields(c.Scope), scope) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "insufficient_scope"})
			return
		}
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
