package tax

import (
	"encoding/json"
	"net/http"

	"github.com/crewdigital/hopper/internal/platform/httpx"
	"github.com/crewdigital/hopper/internal/platform/validator"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for tax operations
type Handler struct {
	taxService *Service
	validator   *validator.Validator
}

// NewHandler creates a new tax handler
func NewHandler(taxService *Service, validator *validator.Validator) *Handler {
	return &Handler{
		taxService: taxService,
		validator:   validator,
	}
}

// ListTaxCategories handles listing all tax categories
func (h *Handler) ListTaxCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.taxService.ListTaxCategories(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list tax categories", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, categories)
}

// ListTaxZones handles listing all tax zones
func (h *Handler) ListTaxZones(w http.ResponseWriter, r *http.Request) {
	zones, err := h.taxService.ListTaxZones(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list tax zones", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, zones)
}

// ListTaxRates handles listing tax rates for a zone
func (h *Handler) ListTaxRates(w http.ResponseWriter, r *http.Request) {
	taxZoneIDStr := r.URL.Query().Get("tax_zone_id")
	taxZoneID, err := uuid.Parse(taxZoneIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_TAX_ZONE_ID", "Invalid tax zone ID", nil)
		return
	}

	rates, err := h.taxService.ListTaxRates(r.Context(), taxZoneID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list tax rates", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, rates)
}

// CalculateTax handles tax calculation
func (h *Handler) CalculateTax(w http.ResponseWriter, r *http.Request) {
	var req CalculateTaxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", validator.GetError(err))
		return
	}

	calc, err := h.taxService.CalculateTax(r.Context(), &req)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to calculate tax", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, calc)
}
