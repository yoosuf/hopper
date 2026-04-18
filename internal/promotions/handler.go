package promotions

import (
	"encoding/json"
	"net/http"

	"github.com/crewdigital/hopper/internal/platform/httpx"
	"github.com/crewdigital/hopper/internal/platform/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles promotion HTTP requests
type Handler struct {
	service   *Service
	validator *validator.Validator
}

// NewHandler creates a new promotions handler
func NewHandler(service *Service, validator *validator.Validator) *Handler {
	return &Handler{
		service:   service,
		validator: validator,
	}
}

// CreatePromotion handles creating a new promotion
// @Summary Create a new promotion
// @Description Create a new promotion/discount/coupon
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreatePromotionRequest true "Promotion details"
// @Success 201 {object} Promotion "Promotion created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /promotions [post]
func (h *Handler) CreatePromotion(w http.ResponseWriter, r *http.Request) {
	var req CreatePromotionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err)
		return
	}

	promotion, err := h.service.CreatePromotion(r.Context(), &req)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "CREATE_PROMOTION_ERROR", "Failed to create promotion", err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, promotion)
}

// GetPromotion handles retrieving a promotion by ID
// @Summary Get promotion by ID
// @Description Retrieve a specific promotion by its ID
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Promotion ID"
// @Success 200 {object} Promotion "Promotion retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid promotion ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Promotion not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /promotions/{id} [get]
func (h *Handler) GetPromotion(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid promotion ID", nil)
		return
	}

	promotion, err := h.service.GetPromotion(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Promotion not found", err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, promotion)
}

// GetPromotionByCode handles retrieving a promotion by code
// @Summary Get promotion by code
// @Description Retrieve a promotion by its code
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Promotion code"
// @Success 200 {object} Promotion "Promotion retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Promotion not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /promotions/code/{code} [get]
func (h *Handler) GetPromotionByCode(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	promotion, err := h.service.GetPromotionByCode(r.Context(), code)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Promotion not found", err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, promotion)
}

// ListPromotions handles listing promotions
// @Summary List promotions
// @Description Retrieve all promotions with optional filters
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param region_id query string false "Filter by region ID"
// @Param restaurant_id query string false "Filter by restaurant ID"
// @Param active_only query boolean false "Filter only active promotions" default(false)
// @Success 200 {array} Promotion "Promotions retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /promotions [get]
func (h *Handler) ListPromotions(w http.ResponseWriter, r *http.Request) {
	regionIDStr := r.URL.Query().Get("region_id")
	restaurantIDStr := r.URL.Query().Get("restaurant_id")
	activeOnly := r.URL.Query().Get("active_only") == "true"

	var regionID, restaurantID *uuid.UUID
	if regionIDStr != "" {
		id, err := uuid.Parse(regionIDStr)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_REGION_ID", "Invalid region ID", nil)
			return
		}
		regionID = &id
	}
	if restaurantIDStr != "" {
		id, err := uuid.Parse(restaurantIDStr)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_RESTAURANT_ID", "Invalid restaurant ID", nil)
			return
		}
		restaurantID = &id
	}

	promotions, err := h.service.ListPromotions(r.Context(), regionID, restaurantID, activeOnly)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "LIST_PROMOTIONS_ERROR", "Failed to list promotions", err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, promotions)
}

// UpdatePromotion handles updating a promotion
// @Summary Update promotion
// @Description Update an existing promotion
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Promotion ID"
// @Param request body UpdatePromotionRequest true "Promotion update details"
// @Success 200 {object} Promotion "Promotion updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Promotion not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /promotions/{id} [put]
func (h *Handler) UpdatePromotion(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid promotion ID", nil)
		return
	}

	var req UpdatePromotionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err)
		return
	}

	promotion, err := h.service.UpdatePromotion(r.Context(), id, &req)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "UPDATE_PROMOTION_ERROR", "Failed to update promotion", err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, promotion)
}

// DeletePromotion handles deleting a promotion
// @Summary Delete promotion
// @Description Delete a promotion
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Promotion ID"
// @Success 204 "Promotion deleted successfully"
// @Failure 400 {object} map[string]string "Invalid promotion ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Promotion not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /promotions/{id} [delete]
func (h *Handler) DeletePromotion(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid promotion ID", nil)
		return
	}

	if err := h.service.DeletePromotion(r.Context(), id); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "DELETE_PROMOTION_ERROR", "Failed to delete promotion", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ValidatePromotion handles validating a promotion code
// @Summary Validate promotion code
// @Description Validate if a promotion code is valid for the given order
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "Promotion validation details"
// @Success 200 {object} map[string]interface{} "Promotion validation result"
// @Failure 400 {object} map[string]string "Invalid request or promotion"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /promotions/validate [post]
func (h *Handler) ValidatePromotion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code         string     `json:"code" validate:"required"`
		OrderValue   int        `json:"order_value" validate:"required,min=0"`
		RegionID     *uuid.UUID `json:"region_id"`
		RestaurantID *uuid.UUID `json:"restaurant_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err)
		return
	}

	promotion, err := h.service.ValidatePromotion(r.Context(), req.Code, req.OrderValue, req.RegionID, req.RestaurantID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_PROMOTION", "Invalid promotion", err)
		return
	}

	response := map[string]interface{}{
		"valid":           true,
		"code":            promotion.Code,
		"type":            promotion.Type,
		"discount_amount": h.service.ApplyPromotion(r.Context(), promotion, req.OrderValue),
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

// UsePromotion handles marking a promotion as used
// @Summary Mark promotion as used
// @Description Mark a promotion as used (decrement usage count)
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "Promotion ID"
// @Success 204 "Promotion marked as used successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Promotion not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /promotions/use [post]
func (h *Handler) UsePromotion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID uuid.UUID `json:"id" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err)
		return
	}

	if err := h.service.UsePromotion(r.Context(), req.ID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "USE_PROMOTION_ERROR", "Failed to use promotion", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
