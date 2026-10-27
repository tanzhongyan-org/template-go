package handler

import "net/http"

// LimitBody caps every request body at maxBytes. Reads past the cap fail with
// *http.MaxBytesError, which decodeJSON turns into a 413 response.
func LimitBody(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
