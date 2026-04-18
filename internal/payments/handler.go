package payments

import (
	"encoding/json"
	"net/http"

	"github.com/crewdigital/hopper/internal/platform/httpx"
	"github.com/crewdigital/hopper/internal/platform/middleware"
	"github.com/crewdigital/hopper/internal/platform/validator"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for payment operations
type Handler struct {
	paymentService *Service
	validator      *validator.Validator
}

// NewHandler creates a new payments handler
func NewHandler(paymentService *Service, validator *validator.Validator) *Handler {
	return &Handler{
		paymentService: paymentService,
		validator:      validator,
	}
}

// CreatePayment handles creating a new payment
// @Summary Create payment
// @Description Create a new payment for an order
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreatePaymentRequest true "Payment details"
// @Success 201 {object} Payment "Payment created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /payments [post]
func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", validator.GetError(err))
		return
	}

	payment, err := h.paymentService.CreatePayment(r.Context(), &req)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create payment", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusCreated, payment)
}

// GetPayment handles getting a payment
// @Summary Get payment by ID
// @Description Retrieve a specific payment by its ID
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Payment ID"
// @Success 200 {object} Payment "Payment retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid payment ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Payment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /payments/{id} [get]
func (h *Handler) GetPayment(w http.ResponseWriter, r *http.Request) {
	paymentIDStr := middleware.URLParam(r, "id")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_PAYMENT_ID", "Invalid payment ID", nil)
		return
	}

	payment, err := h.paymentService.GetPayment(r.Context(), paymentID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get payment", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, payment)
}

// ListOrderPayments handles listing payments for an order
// @Summary List order payments
// @Description Retrieve all payments for a specific order
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order_id path string true "Order ID"
// @Success 200 {array} Payment "Payments retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid order ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /payments/order/{order_id} [get]
func (h *Handler) ListOrderPayments(w http.ResponseWriter, r *http.Request) {
	orderIDStr := middleware.URLParam(r, "order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ORDER_ID", "Invalid order ID", nil)
		return
	}

	payments, err := h.paymentService.ListOrderPayments(r.Context(), orderID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list payments", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, payments)
}

// UpdatePaymentStatus handles updating a payment's status (internal use)
// @Summary Update payment status
// @Description Update the status of a payment (internal use for payment providers)
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Payment ID"
// @Param status query string true "New status"
// @Param provider_id query string false "Payment provider ID"
// @Success 200 {object} map[string]string "Payment status updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Payment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /payments/{id}/status [put]
func (h *Handler) UpdatePaymentStatus(w http.ResponseWriter, r *http.Request) {
	paymentIDStr := middleware.URLParam(r, "id")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_PAYMENT_ID", "Invalid payment ID", nil)
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		httpx.WriteError(w, http.StatusBadRequest, "MISSING_STATUS", "Status is required", nil)
		return
	}

	var providerID *string
	providerIDStr := r.URL.Query().Get("provider_id")
	if providerIDStr != "" {
		providerID = &providerIDStr
	}

	err = h.paymentService.UpdatePaymentStatus(r.Context(), paymentID, status, providerID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update payment status", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Payment status updated successfully",
	})
}
