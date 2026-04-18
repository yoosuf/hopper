package reviews

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles review operations
type Service struct {
	repo Repository
}

// New creates a new reviews service
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateReviewRequest represents a create review request
type CreateReviewRequest struct {
	TargetType string `json:"target_type" validate:"required,oneof=restaurant courier menu_item"`
	TargetID   uuid.UUID `json:"target_id" validate:"required"`
	Rating     int      `json:"rating" validate:"required,min=1,max=5"`
	Comment    string   `json:"comment" validate:"max=1000"`
}

// UpdateReviewRequest represents an update review request
type UpdateReviewRequest struct {
	Rating  int    `json:"rating" validate:"min=1,max=5"`
	Comment string `json:"comment" validate:"max=1000"`
}

// ReviewResponse represents a review response
type ReviewResponse struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	TargetType string    `json:"target_type"`
	TargetID   uuid.UUID `json:"target_id"`
	Rating     int       `json:"rating"`
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateReview creates a new review
func (s *Service) CreateReview(ctx context.Context, userID uuid.UUID, req *CreateReviewRequest) (*ReviewResponse, error) {
	// Check if user already reviewed this target
	existingReview, err := s.repo.GetByUserAndTarget(ctx, userID, req.TargetType, req.TargetID)
	if err == nil && existingReview != nil {
		return nil, fmt.Errorf("user has already reviewed this %s", req.TargetType)
	}

	review := &Review{
		ID:         uuid.New(),
		UserID:     userID,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Rating:     req.Rating,
		Comment:    req.Comment,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, review); err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	return &ReviewResponse{
		ID:         review.ID,
		UserID:     review.UserID,
		TargetType: review.TargetType,
		TargetID:   review.TargetID,
		Rating:     review.Rating,
		Comment:    review.Comment,
		CreatedAt:  review.CreatedAt,
		UpdatedAt:  review.UpdatedAt,
	}, nil
}

// GetReview retrieves a review by ID
func (s *Service) GetReview(ctx context.Context, id uuid.UUID) (*ReviewResponse, error) {
	review, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	return &ReviewResponse{
		ID:         review.ID,
		UserID:     review.UserID,
		TargetType: review.TargetType,
		TargetID:   review.TargetID,
		Rating:     review.Rating,
		Comment:    review.Comment,
		CreatedAt:  review.CreatedAt,
		UpdatedAt:  review.UpdatedAt,
	}, nil
}

// UpdateReview updates a review
func (s *Service) UpdateReview(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *UpdateReviewRequest) (*ReviewResponse, error) {
	review, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	// Check ownership
	if review.UserID != userID {
		return nil, fmt.Errorf("review does not belong to user")
	}

	review.Rating = req.Rating
	review.Comment = req.Comment
	review.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, review); err != nil {
		return nil, fmt.Errorf("failed to update review: %w", err)
	}

	return &ReviewResponse{
		ID:         review.ID,
		UserID:     review.UserID,
		TargetType: review.TargetType,
		TargetID:   review.TargetID,
		Rating:     review.Rating,
		Comment:    review.Comment,
		CreatedAt:  review.CreatedAt,
		UpdatedAt:  review.UpdatedAt,
	}, nil
}

// DeleteReview deletes a review
func (s *Service) DeleteReview(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	review, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get review: %w", err)
	}

	// Check ownership
	if review.UserID != userID {
		return fmt.Errorf("review does not belong to user")
	}

	return s.repo.Delete(ctx, id)
}

// ListTargetReviews lists reviews for a target
func (s *Service) ListTargetReviews(ctx context.Context, targetType string, targetID uuid.UUID, limit int) ([]*ReviewResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // default limit
	}

	reviews, err := s.repo.ListByTarget(ctx, targetType, targetID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list reviews: %w", err)
	}

	var responses []*ReviewResponse
	for _, review := range reviews {
		responses = append(responses, &ReviewResponse{
			ID:         review.ID,
			UserID:     review.UserID,
			TargetType: review.TargetType,
			TargetID:   review.TargetID,
			Rating:     review.Rating,
			Comment:    review.Comment,
			CreatedAt:  review.CreatedAt,
			UpdatedAt:  review.UpdatedAt,
		})
	}

	return responses, nil
}

// ListUserReviews lists reviews by a user
func (s *Service) ListUserReviews(ctx context.Context, userID uuid.UUID, limit int) ([]*ReviewResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // default limit
	}

	reviews, err := s.repo.ListByUser(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list reviews: %w", err)
	}

	var responses []*ReviewResponse
	for _, review := range reviews {
		responses = append(responses, &ReviewResponse{
			ID:         review.ID,
			UserID:     review.UserID,
			TargetType: review.TargetType,
			TargetID:   review.TargetID,
			Rating:     review.Rating,
			Comment:    review.Comment,
			CreatedAt:  review.CreatedAt,
			UpdatedAt:  review.UpdatedAt,
		})
	}

	return responses, nil
}

// GetTargetRatingStats retrieves rating statistics for a target
func (s *Service) GetTargetRatingStats(ctx context.Context, targetType string, targetID uuid.UUID) (float64, int, error) {
	avgRating, count, err := s.repo.GetAverageRating(ctx, targetType, targetID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get rating stats: %w", err)
	}

	return avgRating, count, nil
}
