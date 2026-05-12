package middleware

import (
	"encoding/json"
	"net/http"
)

const (
	APIKeyHeader   = "Api_key"
	ExpectedAPIKey = "apitest"
)

// Auth returns a handler that validates the Api_key header before
// delegating to the next handler. If the key is missing or invalid,
// a 401 Unauthorized JSON response is returned.
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get(APIKeyHeader)
		if key != ExpectedAPIKey {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Unauthorized",
			})
			return
		}
		next.ServeHTTP(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(v)
}