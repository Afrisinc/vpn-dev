package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/ElissaDesign/vpn-dev/internal/utils/response"
)

// PanicRecovery middleware recovers from panics and logs the error
func PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v\n%s", err, debug.Stack())
				w.Header().Set("Content-Type", "application/json")
				response.InternalError(w, "An unexpected error occurred", "internal_error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
