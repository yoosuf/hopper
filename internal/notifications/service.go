package notifications

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service handles notification operations
type Service struct {
	repo Repository
}

// New creates a new notifications service
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// Notification represents a notification
type Notification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      string
	Title     string
	Body      string
	Data      string
	IsRead    bool
	CreatedAt string
}

// CreateNotification creates a new notification
func (s *Service) CreateNotification(ctx context.Context, userID uuid.UUID, notificationType, title, body string, data string) (*Notification, error) {
	notification := &Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      notificationType,
		Title:     title,
		Body:      body,
		Data:      data,
		IsRead:    false,
		CreatedAt: "now",
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return notification, nil
}

// GetNotification retrieves a notification by ID
func (s *Service) GetNotification(ctx context.Context, id uuid.UUID) (*Notification, error) {
	return s.repo.GetByID(ctx, id)
}

// ListUserNotifications lists notifications for a user
func (s *Service) ListUserNotifications(ctx context.Context, userID uuid.UUID) ([]*Notification, error) {
	return s.repo.ListByUser(ctx, userID)
}

// MarkAsRead marks a notification as read
func (s *Service) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	notification, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}

	if notification.UserID != userID {
		return fmt.Errorf("notification does not belong to user")
	}

	return s.repo.MarkAsRead(ctx, id)
}

// MarkAllAsRead marks all notifications for a user as read
func (s *Service) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}
