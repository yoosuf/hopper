package notifications

import (
	"net/http"

	"github.com/crewdigital/hopper/internal/platform/httpx"
	"github.com/crewdigital/hopper/internal/platform/middleware"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for notification operations
type Handler struct {
	notificationService *Service
}

// NewHandler creates a new notifications handler
func NewHandler(notificationService *Service) *Handler {
	return &Handler{
		notificationService: notificationService,
	}
}

// GetNotification handles getting a notification
func (h *Handler) GetNotification(w http.ResponseWriter, r *http.Request) {
	notificationIDStr := middleware.URLParam(r, "id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_NOTIFICATION_ID", "Invalid notification ID", nil)
		return
	}

	notification, err := h.notificationService.GetNotification(r.Context(), notificationID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get notification", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, notification)
}

// ListUserNotifications handles listing notifications for the current user
func (h *Handler) ListUserNotifications(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	notifications, err := h.notificationService.ListUserNotifications(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list notifications", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, notifications)
}

// MarkAsRead handles marking a notification as read
func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	notificationIDStr := middleware.URLParam(r, "id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_NOTIFICATION_ID", "Invalid notification ID", nil)
		return
	}

	err = h.notificationService.MarkAsRead(r.Context(), notificationID, userID)
	if err != nil {
		if err.Error() == "notification does not belong to user" {
			httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Notification does not belong to user", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark notification as read", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Notification marked as read",
	})
}

// MarkAllAsRead handles marking all notifications for the current user as read
func (h *Handler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	err = h.notificationService.MarkAllAsRead(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark all notifications as read", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "All notifications marked as read",
	})
}
