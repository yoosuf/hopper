package admin

import (
	"net/http"

	"github.com/crewdigital/hopper/internal/platform/httpx"
	"github.com/crewdigital/hopper/internal/platform/middleware"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for admin operations
type Handler struct {
	adminService *Service
}

// NewHandler creates a new admin handler
func NewHandler(adminService *Service) *Handler {
	return &Handler{
		adminService: adminService,
	}
}

// ApproveRestaurant handles approving a restaurant
// @Summary Approve restaurant
// @Description Approve a restaurant registration (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Restaurant ID"
// @Success 200 {object} map[string]string "Restaurant approved successfully"
// @Failure 400 {object} map[string]string "Invalid restaurant ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - admin only"
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/restaurants/{id}/approve [put]
func (h *Handler) ApproveRestaurant(w http.ResponseWriter, r *http.Request) {
	restaurantIDStr := middleware.URLParam(r, "id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_RESTAURANT_ID", "Invalid restaurant ID", nil)
		return
	}

	err = h.adminService.ApproveRestaurant(r.Context(), restaurantID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to approve restaurant", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Restaurant approved successfully",
	})
}

// RejectRestaurant handles rejecting a restaurant
// @Summary Reject restaurant
// @Description Reject a restaurant registration (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Restaurant ID"
// @Success 200 {object} map[string]string "Restaurant rejected successfully"
// @Failure 400 {object} map[string]string "Invalid restaurant ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - admin only"
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/restaurants/{id}/reject [put]
func (h *Handler) RejectRestaurant(w http.ResponseWriter, r *http.Request) {
	restaurantIDStr := middleware.URLParam(r, "id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_RESTAURANT_ID", "Invalid restaurant ID", nil)
		return
	}

	err = h.adminService.RejectRestaurant(r.Context(), restaurantID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to reject restaurant", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Restaurant rejected successfully",
	})
}

// ListPendingRestaurants handles listing restaurants pending approval
// @Summary List pending restaurants
// @Description Retrieve all restaurants pending approval (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Pending restaurants retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - admin only"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/restaurants/pending [get]
func (h *Handler) ListPendingRestaurants(w http.ResponseWriter, r *http.Request) {
	restaurantIDs, err := h.adminService.ListPendingRestaurants(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list pending restaurants", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"restaurant_ids": restaurantIDs,
	})
}

// GetSystemStats handles getting system statistics
// @Summary Get system statistics
// @Description Retrieve system statistics (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SystemStats "System statistics retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - admin only"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/stats [get]
func (h *Handler) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.adminService.GetSystemStats(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get system stats", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, stats)
}
