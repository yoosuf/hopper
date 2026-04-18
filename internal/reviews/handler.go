package reviews

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/crewdigital/hopper/internal/platform/httpx"
	"github.com/crewdigital/hopper/internal/platform/middleware"
	"github.com/crewdigital/hopper/internal/platform/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles review HTTP requests
type Handler struct {
	service   *Service
	validator *validator.Validator
}

// NewHandler creates a new reviews handler
func NewHandler(service *Service, validator *validator.Validator) *Handler {
	return &Handler{
		service:   service,
		validator: validator,
	}
}

// CreateReview handles creating a new review
// @Summary Create a new review
// @Description Create a review for a restaurant, courier, or menu item
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateReviewRequest true "Review details"
// @Success 201 {object} ReviewResponse "Review created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 409 {object} map[string]string "Review already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /reviews [post]
func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	var req CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	review, err := h.service.CreateReview(r.Context(), userUUID, &req)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "CREATE_REVIEW_ERROR", "Failed to create review", err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, review)
}

// GetReview handles retrieving a review by ID
// @Summary Get review by ID
// @Description Retrieve a specific review by its ID
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Review ID"
// @Success 200 {object} ReviewResponse "Review retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid review ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Review not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /reviews/{id} [get]
func (h *Handler) GetReview(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid review ID", nil)
		return
	}

	review, err := h.service.GetReview(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Review not found", err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, review)
}

// UpdateReview handles updating a review
// @Summary Update a review
// @Description Update an existing review (only by the reviewer)
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Review ID"
// @Param request body UpdateReviewRequest true "Review update details"
// @Success 200 {object} ReviewResponse "Review updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - not your review"
// @Failure 404 {object} map[string]string "Review not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /reviews/{id} [put]
func (h *Handler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid review ID", nil)
		return
	}

	var req UpdateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	review, err := h.service.UpdateReview(r.Context(), id, userUUID, &req)
	if err != nil {
		httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Failed to update review", err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, review)
}

// DeleteReview handles deleting a review
// @Summary Delete a review
// @Description Delete a review (only by the reviewer)
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Review ID"
// @Success 204 "Review deleted successfully"
// @Failure 400 {object} map[string]string "Invalid review ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - not your review"
// @Failure 404 {object} map[string]string "Review not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /reviews/{id} [delete]
func (h *Handler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid review ID", nil)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	if err := h.service.DeleteReview(r.Context(), id, userUUID); err != nil {
		httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Failed to delete review", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListTargetReviews handles listing reviews for a target
// @Summary List reviews for a target
// @Description Retrieve all reviews for a specific target (restaurant, courier, or menu item)
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param target_type path string true "Target type (restaurant, courier, menu_item)"
// @Param target_id path string true "Target ID"
// @Param limit query int false "Limit number of results" default(20)
// @Success 200 {array} ReviewResponse "Reviews retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /reviews/{target_type}/{target_id} [get]
func (h *Handler) ListTargetReviews(w http.ResponseWriter, r *http.Request) {
	targetType := chi.URLParam(r, "target_type")
	targetIDStr := chi.URLParam(r, "target_id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid target ID", nil)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_LIMIT", "Invalid limit parameter", nil)
			return
		}
	}

	reviews, err := h.service.ListTargetReviews(r.Context(), targetType, targetID, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "LIST_REVIEWS_ERROR", "Failed to list reviews", err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, reviews)
}

// ListUserReviews handles listing reviews by a user
// @Summary List user reviews
// @Description Retrieve all reviews created by the authenticated user
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit number of results" default(20)
// @Success 200 {array} ReviewResponse "Reviews retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /reviews/my [get]
func (h *Handler) ListUserReviews(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	var limit int
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_LIMIT", "Invalid limit parameter", nil)
			return
		}
	} else {
		limit = 20
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	reviews, err := h.service.ListUserReviews(r.Context(), userUUID, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "LIST_REVIEWS_ERROR", "Failed to list reviews", err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, reviews)
}

// GetTargetRatingStats handles retrieving rating statistics for a target
// @Summary Get rating statistics for target
// @Description Retrieve average rating and review count for a specific target
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param target_type path string true "Target type (restaurant, courier, menu_item)"
// @Param target_id path string true "Target ID"
// @Success 200 {object} map[string]interface{} "Rating statistics retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid target ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /reviews/{target_type}/{target_id}/stats [get]
func (h *Handler) GetTargetRatingStats(w http.ResponseWriter, r *http.Request) {
	targetType := chi.URLParam(r, "target_type")
	targetIDStr := chi.URLParam(r, "target_id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid target ID", nil)
		return
	}

	avgRating, count, err := h.service.GetTargetRatingStats(r.Context(), targetType, targetID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "GET_STATS_ERROR", "Failed to get rating stats", err)
		return
	}

	response := map[string]interface{}{
		"average_rating": avgRating,
		"review_count":   count,
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}
