package orders

import (
	"encoding/json"
	"net/http"

	"github.com/crewdigital/hopper/internal/platform/httpx"
	"github.com/crewdigital/hopper/internal/platform/middleware"
	"github.com/crewdigital/hopper/internal/platform/validator"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for order operations
type Handler struct {
	orderService *Service
	validator    *validator.Validator
}

// NewHandler creates a new orders handler
func NewHandler(orderService *Service, validator *validator.Validator) *Handler {
	return &Handler{
		orderService: orderService,
		validator:    validator,
	}
}

// CreateOrder handles creating a new order
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	customerIDStr := middleware.GetUserID(r.Context())
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID", nil)
		return
	}

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", validator.GetError(err))
		return
	}

	order, err := h.orderService.CreateOrder(r.Context(), customerID, &req)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create order", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusCreated, order)
}

// GetOrder handles getting an order
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderIDStr := middleware.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ORDER_ID", "Invalid order ID", nil)
		return
	}

	order, err := h.orderService.GetOrder(r.Context(), orderID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get order", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, order)
}

// ListCustomerOrders handles listing orders for the current customer
func (h *Handler) ListCustomerOrders(w http.ResponseWriter, r *http.Request) {
	customerIDStr := middleware.GetUserID(r.Context())
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID", nil)
		return
	}

	orders, err := h.orderService.ListCustomerOrders(r.Context(), customerID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list orders", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, orders)
}

// ListRestaurantOrders handles listing orders for a restaurant
func (h *Handler) ListRestaurantOrders(w http.ResponseWriter, r *http.Request) {
	restaurantIDStr := middleware.URLParam(r, "restaurant_id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_RESTAURANT_ID", "Invalid restaurant ID", nil)
		return
	}

	orders, err := h.orderService.ListRestaurantOrders(r.Context(), restaurantID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list orders", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, orders)
}

// CancelOrder handles cancelling an order
func (h *Handler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	customerIDStr := middleware.GetUserID(r.Context())
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID", nil)
		return
	}

	orderIDStr := middleware.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ORDER_ID", "Invalid order ID", nil)
		return
	}

	err = h.orderService.CancelOrder(r.Context(), orderID, customerID)
	if err != nil {
		if err.Error() == "order does not belong to customer" {
			httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Order does not belong to customer", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cancel order", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Order cancelled successfully",
	})
}
