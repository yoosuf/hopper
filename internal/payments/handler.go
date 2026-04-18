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
