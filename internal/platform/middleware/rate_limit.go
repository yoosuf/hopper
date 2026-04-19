package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/yoosuf/hopper/internal/platform/cache"
	"github.com/yoosuf/hopper/internal/platform/logger"
)

// DistributedRateLimiter implements distributed rate limiting using cache
type DistributedRateLimiter struct {
	cache           cache.Cache
	requestsPerMinute int
	window          time.Duration
	logger          logger.Logger
}

// NewDistributedRateLimiter creates a new distributed rate limiter
func NewDistributedRateLimiter(cache cache.Cache, requestsPerMinute int, log logger.Logger) *DistributedRateLimiter {
	return &DistributedRateLimiter{
		cache:            cache,
		requestsPerMinute: requestsPerMinute,
		window:           time.Minute,
		logger:           log,
	}
}

// Allow checks if a request should be allowed based on distributed rate limit
func (drl *DistributedRateLimiter) Allow(ctx context.Context, key string) bool {
	cacheKey := fmt.Sprintf("ratelimit:%s", key)
	
	// Get current count from cache
	var count int
	if err := drl.cache.Get(ctx, cacheKey, &count); err != nil {
		// Cache miss - first request
		count = 0
	}

	// Check if under limit
	if count >= drl.requestsPerMinute {
		return false
	}

	// Increment count and set with TTL
	count++
	if err := drl.cache.Set(ctx, cacheKey, count, drl.window); err != nil {
		drl.logger.Error("Failed to set rate limit cache", logger.F("key", cacheKey), logger.F("error", err))
		// Allow request if cache fails (fail open)
		return true
	}

	return true
}

// DistributedRateLimit creates a distributed rate limiting middleware
func DistributedRateLimit(cache cache.Cache, requestsPerMinute int, log logger.Logger) func(next http.Handler) http.Handler {
	limiter := NewDistributedRateLimiter(cache, requestsPerMinute, log)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use IP address as the rate limit key
			key := r.RemoteAddr

			if !limiter.Allow(r.Context(), key) {
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))
				w.Header().Set("X-RateLimit-Remaining", "0")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
