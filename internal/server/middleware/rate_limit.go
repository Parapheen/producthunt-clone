package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/user"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) isAllowed(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get existing requests for this key
	requests, exists := rl.requests[key]
	if !exists {
		requests = []time.Time{}
	}

	// Filter out old requests outside the window
	var validRequests []time.Time
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if we're under the limit
	if len(validRequests) < rl.limit {
		// Add current request
		validRequests = append(validRequests, now)
		rl.requests[key] = validRequests
		return true
	}

	// Update requests (even though we're rejecting)
	rl.requests[key] = validRequests
	return false
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	for key, requests := range rl.requests {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(windowStart) {
				validRequests = append(validRequests, reqTime)
			}
		}
		if len(validRequests) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = validRequests
		}
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limit int, window time.Duration) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(limit, window)

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for range ticker.C {
			limiter.cleanup()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			clientIP := r.RemoteAddr
			if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
				clientIP = forwardedFor
			} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				clientIP = realIP
			}

			// Check rate limit
			if !limiter.isAllowed(clientIP) {
				w.Header().Set("Retry-After", window.String())
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// PerUserRateLimitMiddleware creates a rate limiting middleware based on user ID
func PerUserRateLimitMiddleware(limit int, window time.Duration) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(limit, window)

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for range ticker.C {
			limiter.cleanup()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from context (if authenticated)
			ctxUser := r.Context().Value(user.ContextKeyUser)
			if ctxUser == nil {
				// Fall back to IP-based rate limiting for anonymous users
				clientIP := r.RemoteAddr
				if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
					clientIP = forwardedFor
				} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
					clientIP = realIP
				}

				if !limiter.isAllowed("anon:" + clientIP) {
					w.Header().Set("Retry-After", window.String())
					http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
					return
				}
			} else {
				// Use user ID for rate limiting
				if u, ok := ctxUser.(*user.User); ok {
					if !limiter.isAllowed("user:" + u.ID.String()) {
						w.Header().Set("Retry-After", window.String())
						http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
