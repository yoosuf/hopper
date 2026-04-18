package idempotency

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Service handles idempotency operations
type Service struct {
	ttl time.Duration
}

// New creates a new idempotency service
func New(ttl time.Duration) *Service {
	return &Service{
		ttl: ttl,
	}
}

// Key represents an idempotency key
type Key struct {
	ID     string
	Scope  string
	Hash   string
	Expires time.Time
}

// Response represents a cached response
type Response struct {
	StatusCode int
	Body       json.RawMessage
}

// GenerateKey generates a unique idempotency key
func GenerateKey() string {
	return uuid.New().String()
}

// HashRequest creates a hash of the request body for comparison
func HashRequest(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

// Check checks if an idempotency key exists and returns the cached response if found
func (s *Service) Check(ctx context.Context, key, scope string) (*Response, bool, error) {
	// TODO: Implement database check against idempotency_keys table
	// For now, return not found
	return nil, false, nil
}

// Store stores an idempotency key with its response
func (s *Service) Store(ctx context.Context, key, scope string, requestHash string, response *Response) error {
	// TODO: Implement database insert into idempotency_keys table
	// The record should include:
	// - key (the idempotency key)
	// - scope (the operation scope, e.g., "order.create")
	// - request_hash (hash of request body for validation)
	// - response_code (HTTP status code)
	// - response_body (JSON response body)
	// - expires_at (current time + TTL)
	return nil
}

// ValidateRequest validates that the current request matches the original request
func (s *Service) ValidateRequest(ctx context.Context, key, scope string, currentHash string) (bool, error) {
	// TODO: Retrieve stored request hash and compare with current hash
	// Return true if they match, false otherwise
	return true, nil
}

// Delete removes an idempotency key (e.g., after successful operation)
func (s *Service) Delete(ctx context.Context, key, scope string) error {
	// TODO: Implement database delete from idempotency_keys table
	return nil
}

// CleanupExpired removes expired idempotency keys
func (s *Service) CleanupExpired(ctx context.Context) error {
	// TODO: Implement cleanup of expired keys from idempotency_keys table
	return nil
}
