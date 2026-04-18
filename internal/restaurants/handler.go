package restaurants

import (
	"encoding/json"
	"net/http"

	"github.com/crewdigital/hopper/internal/platform/httpx"
	"github.com/crewdigital/hopper/internal/platform/middleware"
	"github.com/crewdigital/hopper/internal/platform/validator"
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
func (h *Handler) ListRestaurants(w http.ResponseWriter, r *http.Request) {
	regionIDStr := r.URL.Query().Get("region_id")
	regionID, err := uuid.Parse(regionIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REGION_ID", "Invalid region ID", nil)
		return
	}

	restaurants, err := h.restaurantService.ListRestaurants(r.Context(), regionID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list restaurants", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, restaurants)
}

// ListMyRestaurants handles listing restaurants owned by the current user
func (h *Handler) ListMyRestaurants(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	restaurants, err := h.restaurantService.ListMyRestaurants(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list restaurants", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, restaurants)
}

// UpdateRestaurant handles updating a restaurant
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
		if err.Error() == "restaurant does not belong to owner" {
			httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Restaurant does not belong to owner", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update restaurant", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, restaurant)
}

// SetRestaurantHours handles setting restaurant operating hours
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
