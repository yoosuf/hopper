package auth

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/yoosuf/hopper/internal/platform/errors"
	"github.com/yoosuf/hopper/internal/platform/httpx"
	"github.com/yoosuf/hopper/internal/platform/validator"
)

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ,containsany=abcdefghijklmnopqrstuvwxyz,containsany=0123456789,containsany=!@#$%^&*"`
	Role      string `json:"role" validate:"required,oneof=customer restaurant_owner courier"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest represents a logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Handler handles HTTP requests for authentication
type Handler struct {
	authService *Service
	validator   *validator.Validator
}

// NewHandler creates a new auth handler
func NewHandler(authService *Service, validator *validator.Validator) *Handler {
	return &Handler{
		authService: authService,
		validator:   validator,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user with email, password, and role
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} map[string]interface{} "User registered successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 409 {object} map[string]string "Email already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", validator.GetError(err))
		return
	}

	user, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Role, req.FirstName, req.LastName)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok && appErr.Code == apperrors.ErrCodeConflict {
			httpx.WriteError(w, http.StatusConflict, string(appErr.Code), appErr.Message, nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to register user", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusCreated, map[string]interface{}{
		"user": user,
	})
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 403 {object} map[string]string "Account suspended"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", validator.GetError(err))
		return
	}

	tokens, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == ErrInvalidCredentials || err == ErrUserNotFound {
			httpx.WriteError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password", nil)
			return
		}
		if err == ErrUserSuspended {
			httpx.WriteError(w, http.StatusForbidden, "ACCOUNT_SUSPENDED", "Account is suspended", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to login", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, tokens)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} map[string]interface{} "Token refreshed successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid or expired token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", validator.GetError(err))
		return
	}

	tokens, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if err == ErrInvalidToken || err == ErrTokenExpired {
			httpx.WriteError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired refresh token", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to refresh token", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, tokens)
}

// Logout handles user logout
// @Summary Logout user
// @Description Invalidate refresh token on logout
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Refresh token to invalidate"
// @Success 204 "Logout successful"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", validator.GetError(err))
		return
	}

	if err := h.authService.Logout(r.Context(), req.RefreshToken); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to logout", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
