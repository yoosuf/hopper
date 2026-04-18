package delivery

import (
	"net/http"

	"github.com/crewdigital/hopper/internal/platform/httpx"
	"github.com/crewdigital/hopper/internal/platform/middleware"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for delivery operations
type Handler struct {
	deliveryService *Service
}

// NewHandler creates a new delivery handler
func NewHandler(deliveryService *Service) *Handler {
	return &Handler{
		deliveryService: deliveryService,
	}
}

// GetDelivery handles getting a delivery
func (h *Handler) GetDelivery(w http.ResponseWriter, r *http.Request) {
	deliveryIDStr := middleware.URLParam(r, "id")
	deliveryID, err := uuid.Parse(deliveryIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_DELIVERY_ID", "Invalid delivery ID", nil)
		return
	}

	delivery, err := h.deliveryService.GetDelivery(r.Context(), deliveryID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get delivery", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, delivery)
}

// ListCourierDeliveries handles listing deliveries for the current courier
func (h *Handler) ListCourierDeliveries(w http.ResponseWriter, r *http.Request) {
	courierIDStr := middleware.GetUserID(r.Context())
	courierID, err := uuid.Parse(courierIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_COURIER_ID", "Invalid courier ID", nil)
		return
	}

	deliveries, err := h.deliveryService.ListCourierDeliveries(r.Context(), courierID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list deliveries", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, deliveries)
}

// UpdateDeliveryStatus handles updating a delivery's status
func (h *Handler) UpdateDeliveryStatus(w http.ResponseWriter, r *http.Request) {
	deliveryIDStr := middleware.URLParam(r, "id")
	deliveryID, err := uuid.Parse(deliveryIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_DELIVERY_ID", "Invalid delivery ID", nil)
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		httpx.WriteError(w, http.StatusBadRequest, "MISSING_STATUS", "Status is required", nil)
		return
	}

	err = h.deliveryService.UpdateStatus(r.Context(), deliveryID, status)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update delivery status", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Delivery status updated successfully",
	})
}

// MarkPickedUp handles marking a delivery as picked up
func (h *Handler) MarkPickedUp(w http.ResponseWriter, r *http.Request) {
	deliveryIDStr := middleware.URLParam(r, "id")
	deliveryID, err := uuid.Parse(deliveryIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_DELIVERY_ID", "Invalid delivery ID", nil)
		return
	}

	err = h.deliveryService.MarkPickedUp(r.Context(), deliveryID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark delivery as picked up", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Delivery marked as picked up",
	})
}

// MarkDelivered handles marking a delivery as delivered
func (h *Handler) MarkDelivered(w http.ResponseWriter, r *http.Request) {
	deliveryIDStr := middleware.URLParam(r, "id")
	deliveryID, err := uuid.Parse(deliveryIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_DELIVERY_ID", "Invalid delivery ID", nil)
		return
	}

	err = h.deliveryService.MarkDelivered(r.Context(), deliveryID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark delivery as delivered", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Delivery marked as delivered",
	})
}
