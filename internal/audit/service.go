package audit

import (
	"context"

	"github.com/yoosuf/hopper/internal/platform/logger"
)

// Service handles audit logging
type Service struct {
	repo   Repository
	logger logger.Logger
}

// New creates a new audit service
func New(repo Repository, log logger.Logger) *Service {
	return &Service{repo: repo, logger: log}
}

// LogAudit logs an audit event
func (s *Service) LogAudit(ctx context.Context, userID string, action string, resource string, resourceID string, details string) error {
	auditLog := &AuditLog{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    details,
	}

	if err := s.repo.Create(ctx, auditLog); err != nil {
		s.logger.Error("Failed to log audit", logger.F("error", err))
		return err
	}

	return nil
}

// AuditLog represents an audit log entry
type AuditLog struct {
	UserID     string
	Action     string
	Resource   string
	ResourceID string
	Details    string
}
