package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/crewdigital/hopper/internal/auth"
	"github.com/crewdigital/hopper/internal/platform/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

const (
	RequestIDKey = "request_id"
	UserIDKey    = "user_id"
	UserRoleKey  = "user_role"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	ValidateAccessToken(token string) (*auth.Claims, error)
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
		AllowCredentials: true,
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

			// Add user info to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

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
	// Simple ID generation - in production, use proper UUID generation via github.com/google/uuid
	return "req-" + "00000000-0000-0000-0000-000000000000"
}
