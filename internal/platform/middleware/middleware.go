package middleware

import (
	"compress/gzip"
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/yoosuf/hopper/internal/auth"
	"github.com/yoosuf/hopper/internal/platform/logger"
)

const (
	RequestIDKey = "request_id"
	UserIDKey    = "user_id"
	UserRoleKey  = "user_role"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	ValidateAccessToken(token string) (interface{}, error)
}

// RequestID adds a unique request ID to the context
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateID()
		}
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logging logs HTTP requests
func Logging(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Info("HTTP request",
				logger.F("method", r.Method),
				logger.F("path", r.URL.Path),
				logger.F("remote_addr", r.RemoteAddr),
			)
			next.ServeHTTP(w, r)
		})
	}
}

// CORS sets up CORS middleware
func CORS(allowedOrigins, allowedMethods, allowedHeaders []string, maxAge int) func(next http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   allowedMethods,
		AllowedHeaders:   allowedHeaders,
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           maxAge,
	})
}

// Auth validates JWT tokens and adds user info to context
// This is a middleware that requires the auth service to be injected
func Auth(authService AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate JWT token and extract user info
			claims, err := authService.ValidateAccessToken(token)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Type assert to *auth.Claims
			authClaims, ok := claims.(*auth.Claims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), UserIDKey, authClaims.UserID)
			ctx = context.WithValue(ctx, UserRoleKey, authClaims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole checks if the user has the required role
func RequireRole(roles ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(UserRoleKey).(string)
			if !ok {
				http.Error(w, "User role not found in context", http.StatusUnauthorized)
				return
			}

			hasRole := false
			for _, role := range roles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GetUserID retrieves the user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// GetUserRole retrieves the user role from context
func GetUserRole(ctx context.Context) string {
	if userRole, ok := ctx.Value(UserRoleKey).(string); ok {
		return userRole
	}
	return ""
}

// URLParam extracts a URL parameter from the chi context
func URLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func generateID() string {
	return "req-" + uuid.New().String()
}

// RateLimiter is a simple in-memory rate limiter
type RateLimiter struct {
	mu           sync.Mutex
	requests     map[string][]time.Time
	requestLimit int
	window       time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestLimit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests:     make(map[string][]time.Time),
		requestLimit: requestLimit,
		window:       window,
	}
}

// Allow checks if a request should be allowed based on rate limit
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Clean up old requests outside the window
	if requests, exists := rl.requests[key]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) <= rl.window {
				validRequests = append(validRequests, reqTime)
			}
		}
		rl.requests[key] = validRequests
	}

	// Check if under limit
	if len(rl.requests[key]) >= rl.requestLimit {
		return false
	}

	// Add current request
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

// Cleanup removes old entries from the rate limiter map to prevent memory leaks
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, requests := range rl.requests {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) <= rl.window {
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

// RateLimit creates a rate limiting middleware
func RateLimit(requestsPerMinute int) func(next http.Handler) http.Handler {
	limiter := NewRateLimiter(requestsPerMinute, time.Minute)

	// Start periodic cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.Cleanup()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use IP address as the rate limit key
			key := r.RemoteAddr

			if !limiter.Allow(key) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders adds security headers to all responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// MaxBodySize limits the maximum size of request bodies
func MaxBodySize(maxBytes int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// Compression adds gzip compression to responses
func Compression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip encoding
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Don't compress already compressed content types
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "image/") || strings.Contains(accept, "video/") || strings.Contains(accept, "application/zip") {
			next.ServeHTTP(w, r)
			return
		}

		// Create gzip response writer
		gz := gzip.NewWriter(w)
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(&gzipWriter{Writer: gz, ResponseWriter: w}, r)
	})
}

// gzipWriter wraps http.ResponseWriter to support gzip compression
type gzipWriter struct {
	*gzip.Writer
	http.ResponseWriter
}

func (g *gzipWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}

func (g *gzipWriter) WriteHeader(statusCode int) {
	g.ResponseWriter.WriteHeader(statusCode)
}

func (g *gzipWriter) Header() http.Header {
	return g.ResponseWriter.Header()
}
