package middleware

import (
	"net"
	"net/http"
	"sync"

	"github.com/ElissaDesign/vpn-dev/internal/utils/response"
	"golang.org/x/time/rate"
)

// RateLimiter stores rate limiters per IP address
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rps      rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(requestsPerSecond),
		burst:    burst,
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Try X-Forwarded-For header first (for proxied requests)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Get the first IP if multiple are present
		ip, _, _ := net.SplitHostPort(forwarded)
		if ip != "" {
			return ip
		}
	}

	// Try X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		ip, _, _ := net.SplitHostPort(realIP)
		if ip != "" {
			return ip
		}
	}

	// Fall back to RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// getLimiter gets or creates a limiter for the client IP
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rps, rl.burst)
		rl.limiters[ip] = limiter
	}

	return limiter
}

// Middleware returns a rate limiting middleware
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			response.TooManyRequests(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}
