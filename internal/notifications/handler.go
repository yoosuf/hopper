package notifications

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/yoosuf/hopper/internal/platform/errors"
	"github.com/yoosuf/hopper/internal/platform/httpx"
	"github.com/yoosuf/hopper/internal/platform/middleware"
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
// @Summary Get notification by ID
// @Description Retrieve a specific notification by its ID
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} Notification "Notification retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid notification ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - not your notification"
// @Failure 404 {object} map[string]string "Notification not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notifications/{id} [get]
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
// @Summary List user notifications
// @Description Retrieve all notifications for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Notification "Notifications retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notifications [get]
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
// @Summary Mark notification as read
// @Description Mark a notification as read (only by the recipient)
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 204 "Notification marked as read successfully"
// @Failure 400 {object} map[string]string "Invalid notification ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - not your notification"
// @Failure 404 {object} map[string]string "Notification not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notifications/{id}/read [put]
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
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code == errors.ErrCodeForbidden {
			httpx.WriteError(w, http.StatusForbidden, string(appErr.Code), appErr.Message, nil)
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
// @Summary Mark all notifications as read
// @Description Mark all notifications for the authenticated user as read
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "All notifications marked as read successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notifications/read-all [put]
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
