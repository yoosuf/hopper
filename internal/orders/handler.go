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
// @Summary Create a new order
// @Description Create a new order with items, restaurant, and delivery address
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateOrderRequest true "Order details"
// @Success 201 {object} Order "Order created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /orders [post]
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
// @Summary Get order by ID
// @Description Retrieve a specific order by its ID
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} Order "Order retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid order ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Order not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /orders/{id} [get]
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
// @Summary List customer orders
// @Description Retrieve all orders for the authenticated customer
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Order "Orders retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /orders/my [get]
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
// @Summary List restaurant orders
// @Description Retrieve all orders for a specific restaurant
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Success 200 {array} Order "Orders retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid restaurant ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - not your restaurant"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /orders/restaurant/{restaurant_id} [get]
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
// @Summary Cancel an order
// @Description Cancel an order by ID (only allowed in certain workflow states)
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 204 "Order cancelled successfully"
// @Failure 400 {object} map[string]string "Invalid order ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Order cannot be cancelled"
// @Failure 404 {object} map[string]string "Order not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /orders/{id}/cancel [put]
func (h *Handler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	orderIDStr := middleware.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ORDER_ID", "Invalid order ID", nil)
		return
	}

	customerIDStr := middleware.GetUserID(r.Context())
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID", nil)
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

	w.WriteHeader(http.StatusNoContent)
	httpx.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Order cancelled successfully",
	})
}
