package audit

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for audit data access
type Repository interface {
	Create(ctx context.Context, auditLog *AuditLog) error
}

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRepository creates a new audit repository
func NewRepository(db *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

// Create creates a new audit log entry
func (r *RepositoryImpl) Create(ctx context.Context, auditLog *AuditLog) error {
	query := `
		INSERT INTO audit_logs (user_id, action, resource, resource_id, details, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW() AT TIME ZONE 'UTC', NOW() AT TIME ZONE 'UTC')
	`

	_, err := r.db.Exec(ctx, query,
		auditLog.UserID,
		auditLog.Action,
		auditLog.Resource,
		auditLog.ResourceID,
		auditLog.Details,
	)

	return err
}
