package users

import (
	"encoding/json"
	"net/http"

	"github.com/yoosuf/hopper/internal/platform/errors"
	"github.com/yoosuf/hopper/internal/platform/httpx"
	"github.com/yoosuf/hopper/internal/platform/middleware"
	"github.com/yoosuf/hopper/internal/platform/validator"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for user operations
type Handler struct {
	userService *Service
	validator   *validator.Validator
}

// NewHandler creates a new users handler
func NewHandler(userService *Service, validator *validator.Validator) *Handler {
	return &Handler{
		userService: userService,
		validator:   validator,
	}
}

// GetProfile handles getting the current user's profile
// @Summary Get user profile
// @Description Retrieve the authenticated user's profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} User "User profile retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/profile [get]
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	profile, err := h.userService.GetProfile(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get profile", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, profile)
}

// UpdateProfile handles updating the current user's profile
// @Summary Update user profile
// @Description Update the authenticated user's profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateProfileRequest true "Profile update details"
// @Success 200 {object} User "Profile updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/profile [put]
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	profile, err := h.userService.UpdateProfile(r.Context(), userID, &req)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update profile", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, profile)
}

// CreateAddress handles creating a new address
// @Summary Create address
// @Description Create a new address for the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateAddressRequest true "Address details"
// @Success 201 {object} Address "Address created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/addresses [post]
func (h *Handler) CreateAddress(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	var req CreateAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", validator.GetError(err))
		return
	}

	address, err := h.userService.CreateAddress(r.Context(), userID, &req)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create address", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusCreated, address)
}

// ListAddresses handles listing all addresses for the current user
// @Summary List user addresses
// @Description Retrieve all addresses for the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Address "Addresses retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/addresses [get]
func (h *Handler) ListAddresses(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	addresses, err := h.userService.ListAddresses(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list addresses", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, addresses)
}

// UpdateAddress handles updating an address
// @Summary Update address
// @Description Update an existing address for the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Address ID"
// @Param request body CreateAddressRequest true "Address update details"
// @Success 200 {object} Address "Address updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - not your address"
// @Failure 404 {object} map[string]string "Address not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/addresses/{id} [put]
func (h *Handler) UpdateAddress(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	addressIDStr := middleware.URLParam(r, "id")
	addressID, err := uuid.Parse(addressIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ADDRESS_ID", "Invalid address ID", nil)
		return
	}

	var req CreateAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", validator.GetError(err))
		return
	}

	address, err := h.userService.UpdateAddress(r.Context(), addressID, userID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code == errors.ErrCodeForbidden {
			httpx.WriteError(w, http.StatusForbidden, string(appErr.Code), appErr.Message, nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update address", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, address)
}

// DeleteAddress handles deleting an address
// @Summary Delete address
// @Description Delete an address for the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Address ID"
// @Success 200 {object} map[string]string "Address deleted successfully"
// @Failure 400 {object} map[string]string "Invalid address ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - not your address"
// @Failure 404 {object} map[string]string "Address not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/addresses/{id} [delete]
func (h *Handler) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	addressIDStr := middleware.URLParam(r, "id")
	addressID, err := uuid.Parse(addressIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ADDRESS_ID", "Invalid address ID", nil)
		return
	}

	err = h.userService.DeleteAddress(r.Context(), addressID, userID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code == errors.ErrCodeForbidden {
			httpx.WriteError(w, http.StatusForbidden, string(appErr.Code), appErr.Message, nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete address", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Address deleted successfully",
	})
}
