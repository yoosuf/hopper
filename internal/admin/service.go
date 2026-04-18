package admin

import (
	"context"

	"github.com/google/uuid"
)

// Service handles admin operations
type Service struct {
	repo Repository
}

// New creates a new admin service
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// ApproveRestaurant approves a restaurant
func (s *Service) ApproveRestaurant(ctx context.Context, restaurantID uuid.UUID) error {
	return s.repo.ApproveRestaurant(ctx, restaurantID)
}

// RejectRestaurant rejects a restaurant
func (s *Service) RejectRestaurant(ctx context.Context, restaurantID uuid.UUID) error {
	return s.repo.RejectRestaurant(ctx, restaurantID)
}

// ListPendingRestaurants lists restaurants pending approval
func (s *Service) ListPendingRestaurants(ctx context.Context) ([]uuid.UUID, error) {
	return s.repo.ListPendingRestaurants(ctx)
}

// GetSystemStats retrieves system statistics
func (s *Service) GetSystemStats(ctx context.Context) (*SystemStats, error) {
	return s.repo.GetSystemStats(ctx)
}

// SystemStats represents system statistics
type SystemStats struct {
	TotalUsers       int
	TotalRestaurants int
	TotalOrders      int
	ActiveDeliveries int
}
