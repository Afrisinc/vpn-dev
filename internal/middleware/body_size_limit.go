package middleware

import (
	"net/http"
)

// BodySizeLimit middleware limits the request body size
func BodySizeLimit(maxSize int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)

			next.ServeHTTP(w, r)
		})
	}
}

// LimitBodySize is a helper to create the middleware with a specific size
func LimitBodySize(size int64) func(http.Handler) http.Handler {
	return BodySizeLimit(size)
}
