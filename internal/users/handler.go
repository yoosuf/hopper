package users

import (
	"encoding/json"
	"net/http"

	"github.com/crewdigital/hopper/internal/platform/httpx"
	"github.com/crewdigital/hopper/internal/platform/middleware"
	"github.com/crewdigital/hopper/internal/platform/validator"
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
		if err.Error() == "address does not belong to user" {
			httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Address does not belong to user", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update address", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, address)
}

// DeleteAddress handles deleting an address
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
		if err.Error() == "address does not belong to user" {
			httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Address does not belong to user", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete address", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Address deleted successfully",
	})
}
