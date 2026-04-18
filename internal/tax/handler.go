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
	validator  *validator.Validator
}

// NewHandler creates a new tax handler
func NewHandler(taxService *Service, validator *validator.Validator) *Handler {
	return &Handler{
		taxService: taxService,
		validator:  validator,
	}
}

// ListTaxCategories handles listing all tax categories
// @Summary List tax categories
// @Description Retrieve all tax categories
// @Tags tax
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} TaxCategory "Tax categories retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /tax/categories [get]
func (h *Handler) ListTaxCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.taxService.ListTaxCategories(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list tax categories", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, categories)
}

// ListTaxZones handles listing all tax zones
// @Summary List tax zones
// @Description Retrieve all tax zones
// @Tags tax
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} TaxZone "Tax zones retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /tax/zones [get]
func (h *Handler) ListTaxZones(w http.ResponseWriter, r *http.Request) {
	zones, err := h.taxService.ListTaxZones(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list tax zones", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, zones)
}

// ListTaxRates handles listing tax rates for a zone
// @Summary List tax rates for zone
// @Description Retrieve all tax rates for a specific tax zone
// @Tags tax
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tax_zone_id query string true "Tax Zone ID"
// @Success 200 {array} TaxRate "Tax rates retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid tax zone ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /tax/rates [get]
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
// @Summary Calculate tax
// @Description Calculate tax for a given order based on region and items
// @Tags tax
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CalculateTaxRequest true "Tax calculation request"
// @Success 200 {object} TaxCalculation "Tax calculated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /tax/calculate [post]
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
