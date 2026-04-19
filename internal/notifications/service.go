package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yoosuf/hopper/internal/platform/logger"
)

// NotificationChannel represents the delivery channel for a notification
type NotificationChannel string

const (
	ChannelPush  NotificationChannel = "push"
	ChannelEmail NotificationChannel = "email"
	ChannelSMS   NotificationChannel = "sms"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	TypeOrderCreated    NotificationType = "order_created"
	TypeOrderConfirmed  NotificationType = "order_confirmed"
	TypeOrderPreparing  NotificationType = "order_preparing"
	TypeOrderReady      NotificationType = "order_ready"
	TypeOrderPickedUp   NotificationType = "order_picked_up"
	TypeOrderDelivered  NotificationType = "order_delivered"
	TypeOrderCancelled  NotificationType = "order_cancelled"
	TypePaymentReceived NotificationType = "payment_received"
	TypePaymentFailed   NotificationType = "payment_failed"
)

// Notification represents a notification
type Notification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      NotificationType
	Title     string
	Body      string
	Data      string
	Channels  []NotificationChannel
	IsRead    bool
	CreatedAt time.Time
}

// NotificationPreference represents user notification preferences
type NotificationPreference struct {
	UserID        uuid.UUID
	Channel       NotificationChannel
	Enabled       bool
	OrderEvents   bool
	PaymentEvents bool
}

// Sender defines the interface for sending notifications
type Sender interface {
	SendPush(ctx context.Context, userID uuid.UUID, title, body string, data map[string]interface{}) error
	SendEmail(ctx context.Context, to string, subject, body string) error
	SendSMS(ctx context.Context, to, message string) error
}

// Service handles notification operations
type Service struct {
	repo   Repository
	sender Sender
	logger logger.Logger
}

// New creates a new notifications service
func New(repo Repository, sender Sender, log logger.Logger) *Service {
	return &Service{repo: repo, sender: sender, logger: log}
}

// CreateNotification creates a new notification and sends it via configured channels
func (s *Service) CreateNotification(ctx context.Context, userID uuid.UUID, notificationType NotificationType, title, body string, channels []NotificationChannel, data map[string]interface{}) (*Notification, error) {
	// Get user preferences for each channel
	enabledChannels := s.getEnabledChannels(ctx, userID, channels)

	notification := &Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      notificationType,
		Title:     title,
		Body:      body,
		Data:      fmt.Sprintf("%v", data),
		Channels:  enabledChannels,
		IsRead:    false,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Send notifications via enabled channels
	if err := s.sendNotifications(ctx, notification, data); err != nil {
		return nil, fmt.Errorf("failed to send notifications: %w", err)
	}

	return notification, nil
}

// getEnabledChannels filters channels based on user preferences
func (s *Service) getEnabledChannels(ctx context.Context, userID uuid.UUID, channels []NotificationChannel) []NotificationChannel {
	var enabled []NotificationChannel
	for _, channel := range channels {
		pref, err := s.repo.GetPreference(ctx, userID, channel)
		if err == nil && pref != nil && pref.Enabled {
			enabled = append(enabled, channel)
		} else if err != nil {
			// If no preference exists, enable by default
			enabled = append(enabled, channel)
		}
	}
	return enabled
}

// sendNotifications sends notifications via configured channels
func (s *Service) sendNotifications(ctx context.Context, notification *Notification, data map[string]interface{}) error {
	for _, channel := range notification.Channels {
		switch channel {
		case ChannelPush:
			if s.sender != nil {
				if err := s.sender.SendPush(ctx, notification.UserID, notification.Title, notification.Body, data); err != nil {
					// Log error but don't fail the entire operation
					s.logger.Error("Failed to send push notification", logger.F("error", err))
				}
			}
		case ChannelEmail:
			if s.sender != nil {
				// Get user email from repository or context
				email := "user@example.com" // TODO: Get user email
				if err := s.sender.SendEmail(ctx, email, notification.Title, notification.Body); err != nil {
					s.logger.Error("Failed to send email notification", logger.F("error", err))
				}
			}
		case ChannelSMS:
			if s.sender != nil {
				// Get user phone from repository or context
				phone := "+1234567890" // TODO: Get user phone
				if err := s.sender.SendSMS(ctx, phone, notification.Body); err != nil {
					s.logger.Error("Failed to send SMS notification", logger.F("error", err))
				}
			}
		}
	}
	return nil
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

// SetNotificationPreference sets user notification preference
func (s *Service) SetNotificationPreference(ctx context.Context, userID uuid.UUID, channel NotificationChannel, enabled bool, orderEvents, paymentEvents bool) error {
	pref := &NotificationPreference{
		UserID:        userID,
		Channel:       channel,
		Enabled:       enabled,
		OrderEvents:   orderEvents,
		PaymentEvents: paymentEvents,
	}
	return s.repo.SetPreference(ctx, pref)
}

// GetNotificationPreference retrieves user notification preference
func (s *Service) GetNotificationPreference(ctx context.Context, userID uuid.UUID, channel NotificationChannel) (*NotificationPreference, error) {
	return s.repo.GetPreference(ctx, userID, channel)
}

// ListNotificationPreferences retrieves all user notification preferences
func (s *Service) ListNotificationPreferences(ctx context.Context, userID uuid.UUID) ([]*NotificationPreference, error) {
	return s.repo.ListPreferences(ctx, userID)
}
