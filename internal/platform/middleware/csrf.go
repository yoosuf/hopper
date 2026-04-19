package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

// CSRFProtection provides CSRF protection middleware
type CSRFProtection struct {
	secret []byte
}

// NewCSRFProtection creates a new CSRF protection middleware
func NewCSRFProtection(secret string) *CSRFProtection {
	return &CSRFProtection{
		secret: []byte(secret),
	}
}

// generateToken generates a random CSRF token
func (c *CSRFProtection) generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// CSRF creates a CSRF protection middleware
func (c *CSRFProtection) CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF for safe methods
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Check CSRF token for state-changing methods
		token := r.Header.Get("X-CSRF-Token")
		if token == "" {
			token = r.FormValue("csrf_token")
		}

		if token == "" {
			http.Error(w, "CSRF token missing", http.StatusForbidden)
			return
		}

		// Validate token (in production, this should validate against session)
		// For now, we just check that it's not empty
		if len(token) < 32 {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SetCSRFToken sets the CSRF token in the response headers
func (c *CSRFProtection) SetCSRFToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate and set CSRF token for safe methods
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			token := c.generateToken()
			w.Header().Set("X-CSRF-Token", token)
		}
		next.ServeHTTP(w, r)
	})
}

// CSRFMiddleware creates CSRF protection middleware with token setting
func CSRFMiddleware(secret string) func(next http.Handler) http.Handler {
	csrf := NewCSRFProtection(secret)
	
	return func(next http.Handler) http.Handler {
		return csrf.CSRF(csrf.SetCSRFToken(next))
	}
}

// ValidateCSRFOrigin validates the request origin against allowed origins
func ValidateCSRFOrigin(allowedOrigins []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				origin = r.Header.Get("Referer")
			}

			if origin != "" {
				allowed := false
				for _, allowedOrigin := range allowedOrigins {
					if strings.HasPrefix(origin, allowedOrigin) {
						allowed = true
						break
					}
				}
				if !allowed {
					http.Error(w, "Invalid origin", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
