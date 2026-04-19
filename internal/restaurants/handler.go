package restaurants

import (
	"encoding/json"
	"net/http"
	"strconv"

	apperrors "github.com/yoosuf/hopper/internal/platform/errors"
	"github.com/yoosuf/hopper/internal/platform/httpx"
	"github.com/yoosuf/hopper/internal/platform/middleware"
	"github.com/yoosuf/hopper/internal/platform/validator"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for restaurant operations
type Handler struct {
	restaurantService *Service
	validator         *validator.Validator
}

// NewHandler creates a new restaurants handler
func NewHandler(restaurantService *Service, validator *validator.Validator) *Handler {
	return &Handler{
		restaurantService: restaurantService,
		validator:         validator,
	}
}

// CreateRestaurant handles creating a new restaurant
// @Summary Create a new restaurant
// @Description Create a new restaurant for the authenticated user
// @Tags restaurants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRestaurantRequest true "Restaurant details"
// @Success 201 {object} Restaurant "Restaurant created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /restaurants [post]
func (h *Handler) CreateRestaurant(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	var req CreateRestaurantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", validator.GetError(err))
		return
	}

	restaurant, err := h.restaurantService.CreateRestaurant(r.Context(), userID, &req)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create restaurant", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusCreated, restaurant)
}

// GetRestaurant handles getting a restaurant
// @Summary Get restaurant by ID
// @Description Retrieve a specific restaurant by its ID
// @Tags restaurants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Restaurant ID"
// @Success 200 {object} Restaurant "Restaurant retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid restaurant ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /restaurants/{id} [get]
func (h *Handler) GetRestaurant(w http.ResponseWriter, r *http.Request) {
	restaurantIDStr := middleware.URLParam(r, "id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_RESTAURANT_ID", "Invalid restaurant ID", nil)
		return
	}

	restaurant, err := h.restaurantService.GetRestaurant(r.Context(), restaurantID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get restaurant", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, restaurant)
}

// ListRestaurants handles listing restaurants for a region
// @Summary List restaurants
// @Description Retrieve all restaurants in a specific region
// @Tags restaurants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param region_id query string true "Region ID"
// @Success 200 {array} Restaurant "Restaurants retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid region ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /restaurants [get]
func (h *Handler) ListRestaurants(w http.ResponseWriter, r *http.Request) {
	regionIDStr := r.URL.Query().Get("region_id")
	regionID, err := uuid.Parse(regionIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REGION_ID", "Invalid region ID", nil)
		return
	}

	// Extract pagination parameters
	limit := 50
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	restaurants, err := h.restaurantService.ListRestaurantsInRegion(r.Context(), regionID, limit, offset)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list restaurants", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, restaurants)
}

// ListMyRestaurants handles listing restaurants owned by the current user
// @Summary List my restaurants
// @Description Retrieve all restaurants owned by the authenticated user
// @Tags restaurants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Restaurant "Restaurants retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /restaurants/my [get]
func (h *Handler) ListMyRestaurants(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	// Extract pagination parameters
	limit := 50
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	restaurants, err := h.restaurantService.ListOwnerRestaurants(r.Context(), userID, limit, offset)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list restaurants", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, restaurants)
}

// UpdateRestaurant handles updating a restaurant
// @Summary Update restaurant
// @Description Update restaurant details (only by owner)
// @Tags restaurants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Restaurant ID"
// @Param request body CreateRestaurantRequest true "Restaurant details"
// @Success 200 {object} Restaurant "Restaurant updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /restaurants/{id} [put]
func (h *Handler) UpdateRestaurant(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	restaurantIDStr := middleware.URLParam(r, "id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_RESTAURANT_ID", "Invalid restaurant ID", nil)
		return
	}

	var req CreateRestaurantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", validator.GetError(err))
		return
	}

	restaurant, err := h.restaurantService.UpdateRestaurant(r.Context(), restaurantID, userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok && appErr.Code == apperrors.ErrCodeForbidden {
			httpx.WriteError(w, http.StatusForbidden, string(appErr.Code), appErr.Message, nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update restaurant", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, restaurant)
}

// SetRestaurantHours handles setting restaurant operating hours
// @Summary Set restaurant operating hours
// @Description Set the operating hours for a restaurant (only by owner)
// @Tags restaurants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Restaurant ID"
// @Param request body []RestaurantHour true "Operating hours"
// @Success 200 {array} RestaurantHour "Operating hours set successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - not your restaurant"
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /restaurants/{id}/hours [put]
func (h *Handler) SetRestaurantHours(w http.ResponseWriter, r *http.Request) {
	restaurantIDStr := middleware.URLParam(r, "id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_RESTAURANT_ID", "Invalid restaurant ID", nil)
		return
	}

	var hours []*RestaurantHour
	if err := json.NewDecoder(r.Body).Decode(&hours); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.restaurantService.SetRestaurantHours(r.Context(), restaurantID, hours); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to set restaurant hours", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Restaurant hours updated successfully",
	})
}

// GetRestaurantHours handles getting restaurant operating hours
// @Summary Get restaurant operating hours
// @Description Retrieve the operating hours for a restaurant
// @Tags restaurants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Restaurant ID"
// @Success 200 {array} RestaurantHour "Operating hours retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid restaurant ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /restaurants/{id}/hours [get]
func (h *Handler) GetRestaurantHours(w http.ResponseWriter, r *http.Request) {
	restaurantIDStr := middleware.URLParam(r, "id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_RESTAURANT_ID", "Invalid restaurant ID", nil)
		return
	}

	hours, err := h.restaurantService.GetRestaurantHours(r.Context(), restaurantID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get restaurant hours", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, hours)
}
