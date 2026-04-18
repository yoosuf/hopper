package idempotency

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for idempotency data access
type Repository interface {
	Get(ctx context.Context, key, scope string) (*Key, *Response, error)
	Create(ctx context.Context, key *Key, response *Response) error
	Delete(ctx context.Context, key, scope string) error
	CleanupExpired(ctx context.Context) error
}

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRepository creates a new idempotency repository
func NewRepository(db *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

// Get retrieves an idempotency key and its response
func (r *RepositoryImpl) Get(ctx context.Context, key, scope string) (*Key, *Response, error) {
	query := `
		SELECT key, scope, request_hash, response_code, response_body, expires_at
		FROM idempotency_keys
		WHERE key = $1 AND scope = $2 AND expires_at > NOW() AT TIME ZONE 'UTC'
	`

	var k Key
	var resp Response
	var expiresAt time.Time

	err := r.db.QueryRow(ctx, query, key, scope).Scan(
		&k.ID,
		&k.Scope,
		&k.Hash,
		&resp.StatusCode,
		&resp.Body,
		&expiresAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	k.Expires = expiresAt
	return &k, &resp, nil
}

// Create stores a new idempotency key with its response
func (r *RepositoryImpl) Create(ctx context.Context, key *Key, response *Response) error {
	query := `
		INSERT INTO idempotency_keys (key, scope, request_hash, response_code, response_body, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW() AT TIME ZONE 'UTC')
	`

	_, err := r.db.Exec(ctx, query,
		key.ID,
		key.Scope,
		key.Hash,
		response.StatusCode,
		response.Body,
		key.Expires,
	)
	return err
}

// Delete removes an idempotency key
func (r *RepositoryImpl) Delete(ctx context.Context, key, scope string) error {
	query := `
		DELETE FROM idempotency_keys
		WHERE key = $1 AND scope = $2
	`

	_, err := r.db.Exec(ctx, query, key, scope)
	return err
}

// CleanupExpired removes expired idempotency keys
func (r *RepositoryImpl) CleanupExpired(ctx context.Context) error {
	query := `
		DELETE FROM idempotency_keys
		WHERE expires_at <= NOW() AT TIME ZONE 'UTC'
	`

	_, err := r.db.Exec(ctx, query)
	return err
}

// Service handles idempotency operations
type Service struct {
	repo Repository
	ttl  time.Duration
}

// New creates a new idempotency service
func New(repo Repository, ttl time.Duration) *Service {
	return &Service{
		repo: repo,
		ttl:  ttl,
	}
}

// Key represents an idempotency key
type Key struct {
	ID      string
	Scope   string
	Hash    string
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
	_, response, err := s.repo.Get(ctx, key, scope)
	if err != nil {
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	return response, true, nil
}

// Store stores an idempotency key with its response
func (s *Service) Store(ctx context.Context, key, scope string, requestHash string, response *Response) error {
	k := &Key{
		ID:      key,
		Scope:   scope,
		Hash:    requestHash,
		Expires: time.Now().UTC().Add(s.ttl),
	}
	return s.repo.Create(ctx, k, response)
}

// ValidateRequest validates that the current request matches the original request
func (s *Service) ValidateRequest(ctx context.Context, key, scope string, currentHash string) (bool, error) {
	storedKey, _, err := s.repo.Get(ctx, key, scope)
	if err != nil {
		return false, err
	}
	if storedKey == nil {
		return false, nil
	}
	return storedKey.Hash == currentHash, nil
}

// Delete removes an idempotency key (e.g., after successful operation)
func (s *Service) Delete(ctx context.Context, key, scope string) error {
	return s.repo.Delete(ctx, key, scope)
}

// CleanupExpired removes expired idempotency keys
func (s *Service) CleanupExpired(ctx context.Context) error {
	return s.repo.CleanupExpired(ctx)
}
