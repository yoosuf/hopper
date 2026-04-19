package delivery

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yoosuf/hopper/internal/notifications"
)

// NotificationAdapter adapts notifications.Service to CourierNotifier.
type NotificationAdapter struct {
	notificationService *notifications.Service
}

// NewNotificationAdapter creates a courier notifier backed by notifications service.
func NewNotificationAdapter(notificationService *notifications.Service) *NotificationAdapter {
	return &NotificationAdapter{notificationService: notificationService}
}

// NotifyCourierAssigned sends a courier assignment notification.
func (a *NotificationAdapter) NotifyCourierAssigned(ctx context.Context, courierID, deliveryID uuid.UUID, etaMinutes int) error {
	if a.notificationService == nil {
		return nil
	}

	_, err := a.notificationService.CreateNotification(
		ctx,
		courierID,
		notifications.TypeDeliveryAssigned,
		"New delivery assigned",
		fmt.Sprintf("A new delivery has been assigned to you. ETA to pickup: %d minutes", etaMinutes),
		[]notifications.NotificationChannel{notifications.ChannelPush, notifications.ChannelSMS},
		map[string]interface{}{
			"delivery_id": deliveryID.String(),
			"eta_minutes": etaMinutes,
		},
	)
	return err
}
