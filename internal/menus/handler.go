package menus

import (
	"encoding/json"
	"net/http"

	"github.com/crewdigital/hopper/internal/platform/httpx"
	"github.com/crewdigital/hopper/internal/platform/middleware"
	"github.com/crewdigital/hopper/internal/platform/validator"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for menu operations
type Handler struct {
	menuService *Service
	validator   *validator.Validator
}

// NewHandler creates a new menus handler
func NewHandler(menuService *Service, validator *validator.Validator) *Handler {
	return &Handler{
		menuService: menuService,
		validator:   validator,
	}
}

// CreateMenuItem handles creating a new menu item
func (h *Handler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	restaurantIDStr := middleware.URLParam(r, "restaurant_id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_RESTAURANT_ID", "Invalid restaurant ID", nil)
		return
	}

	var req CreateMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", validator.GetError(err))
		return
	}

	menuItem, err := h.menuService.CreateMenuItem(r.Context(), restaurantID, &req)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create menu item", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusCreated, menuItem)
}

// GetMenuItem handles getting a menu item
func (h *Handler) GetMenuItem(w http.ResponseWriter, r *http.Request) {
	menuItemIDStr := middleware.URLParam(r, "id")
	menuItemID, err := uuid.Parse(menuItemIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_MENU_ITEM_ID", "Invalid menu item ID", nil)
		return
	}

	menuItem, err := h.menuService.GetMenuItem(r.Context(), menuItemID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get menu item", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, menuItem)
}

// ListMenuItems handles listing menu items for a restaurant
func (h *Handler) ListMenuItems(w http.ResponseWriter, r *http.Request) {
	restaurantIDStr := middleware.URLParam(r, "restaurant_id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_RESTAURANT_ID", "Invalid restaurant ID", nil)
		return
	}

	menuItems, err := h.menuService.ListMenuItems(r.Context(), restaurantID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list menu items", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, menuItems)
}

// UpdateMenuItem handles updating a menu item
func (h *Handler) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	restaurantIDStr := middleware.URLParam(r, "restaurant_id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_RESTAURANT_ID", "Invalid restaurant ID", nil)
		return
	}

	menuItemIDStr := middleware.URLParam(r, "id")
	menuItemID, err := uuid.Parse(menuItemIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_MENU_ITEM_ID", "Invalid menu item ID", nil)
		return
	}

	var req CreateMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", validator.GetError(err))
		return
	}

	menuItem, err := h.menuService.UpdateMenuItem(r.Context(), menuItemID, restaurantID, &req)
	if err != nil {
		if err.Error() == "menu item does not belong to restaurant" {
			httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Menu item does not belong to restaurant", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update menu item", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, menuItem)
}

// DeleteMenuItem handles deleting a menu item
func (h *Handler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	restaurantIDStr := middleware.URLParam(r, "restaurant_id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_RESTAURANT_ID", "Invalid restaurant ID", nil)
		return
	}

	menuItemIDStr := middleware.URLParam(r, "id")
	menuItemID, err := uuid.Parse(menuItemIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_MENU_ITEM_ID", "Invalid menu item ID", nil)
		return
	}

	err = h.menuService.DeleteMenuItem(r.Context(), menuItemID, restaurantID)
	if err != nil {
		if err.Error() == "menu item does not belong to restaurant" {
			httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Menu item does not belong to restaurant", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete menu item", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Menu item deleted successfully",
	})
}
