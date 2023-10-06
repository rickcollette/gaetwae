package ratelimit

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

var visitors = make(map[string]*rate.Limiter)
var mtx sync.Mutex

// Create a new rate limiter for each new IP
func getVisitor(ip string, r *rate.Limiter) *rate.Limiter {
	mtx.Lock()
	defer mtx.Unlock()

	visitor, exists := visitors[ip]
	if !exists {
		visitors[ip] = r
		return r
	}

	return visitor
}

// Middleware for rate limiting based on IP
func RateLimitMiddleware(r *rate.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			limiter := getVisitor(req.RemoteAddr, r)
			
			if !limiter.Allow() {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}
