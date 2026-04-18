package audit

import (
	"context"
	"log"
)

// Service handles audit logging
type Service struct {
	repo Repository
}

// New creates a new audit service
func New(repo Repository) *Service {
	return &Service{repo: repo}
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
		log.Printf("Failed to log audit: %v", err)
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
