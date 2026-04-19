package menus

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/yoosuf/hopper/internal/platform/errors"
	"github.com/yoosuf/hopper/internal/platform/httpx"
	"github.com/yoosuf/hopper/internal/platform/middleware"
	"github.com/yoosuf/hopper/internal/platform/validator"
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
// @Summary Create menu item
// @Description Create a new menu item for a restaurant
// @Tags menus
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Param request body CreateMenuItemRequest true "Menu item details"
// @Success 201 {object} MenuItem "Menu item created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - not your restaurant"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /menus/{restaurant_id}/items [post]
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
// @Summary Get menu item by ID
// @Description Retrieve a specific menu item by its ID
// @Tags menus
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Menu Item ID"
// @Success 200 {object} MenuItem "Menu item retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid menu item ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Menu item not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /menus/items/{id} [get]
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
// @Summary List menu items for restaurant
// @Description Retrieve all menu items for a specific restaurant
// @Tags menus
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Success 200 {array} MenuItem "Menu items retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid restaurant ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /menus/{restaurant_id}/items [get]
func (h *Handler) ListMenuItems(w http.ResponseWriter, r *http.Request) {
	restaurantIDStr := middleware.URLParam(r, "restaurant_id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_RESTAURANT_ID", "Invalid restaurant ID", nil)
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

	menuItems, err := h.menuService.ListMenuItems(r.Context(), restaurantID, limit, offset)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list menu items", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, menuItems)
}

// UpdateMenuItem handles updating a menu item
// @Summary Update menu item
// @Description Update a menu item (only by restaurant owner)
// @Tags menus
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Param id path string true "Menu Item ID"
// @Param request body CreateMenuItemRequest true "Menu item details"
// @Success 200 {object} MenuItem "Menu item updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - not your restaurant"
// @Failure 404 {object} map[string]string "Menu item not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /menus/{restaurant_id}/items/{id} [put]
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
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code == errors.ErrCodeForbidden {
			httpx.WriteError(w, http.StatusForbidden, string(appErr.Code), appErr.Message, nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update menu item", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, menuItem)
}

// DeleteMenuItem handles deleting a menu item
// @Summary Delete menu item
// @Description Delete a menu item (only by restaurant owner)
// @Tags menus
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Param id path string true "Menu Item ID"
// @Success 204 "Menu item deleted successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - not your restaurant"
// @Failure 404 {object} map[string]string "Menu item not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /menus/{restaurant_id}/items/{id} [delete]
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

	if err := h.menuService.DeleteMenuItem(r.Context(), menuItemID, restaurantID); err != nil {
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code == errors.ErrCodeForbidden {
			httpx.WriteError(w, http.StatusForbidden, string(appErr.Code), appErr.Message, nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete menu item", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
