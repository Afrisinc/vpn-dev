package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// ResponseWriter wraps http.ResponseWriter to capture status code
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLogger middleware logs HTTP requests with structured logging
func RequestLogger(logger *zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer
			wrapped := &ResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Get request ID from context
			requestID := GetRequestID(r.Context())

			// Serve request
			next.ServeHTTP(wrapped, r)

			// Log request
			duration := time.Since(start)
			logger.Info().
				Str("requestId", requestID).
				Str("method", r.Method).
				Str("path", r.RequestURI).
				Int("statusCode", wrapped.statusCode).
				Dur("duration", duration).
				Str("remoteAddr", r.RemoteAddr).
				Msg("HTTP request")
		})
	}
}
