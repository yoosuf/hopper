package menus

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service handles menu operations
type Service struct {
	repo Repository
}

// New creates a new menus service
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// MenuItem represents a menu item
type MenuItem struct {
	ID            uuid.UUID
	RestaurantID  uuid.UUID
	Name          string
	Description   string
	Category      string
	BasePrice     int
	TaxCategoryID uuid.UUID
	IsActive      bool
}

// CreateMenuItemRequest represents a create menu item request
type CreateMenuItemRequest struct {
	Name          string
	Description   string
	Category      string
	BasePrice     int
	TaxCategoryID uuid.UUID
}

// CreateMenuItem creates a new menu item
func (s *Service) CreateMenuItem(ctx context.Context, restaurantID uuid.UUID, req *CreateMenuItemRequest) (*MenuItem, error) {
	menuItem := &MenuItem{
		ID:            uuid.New(),
		RestaurantID:  restaurantID,
		Name:          req.Name,
		Description:   req.Description,
		Category:      req.Category,
		BasePrice:     req.BasePrice,
		TaxCategoryID: req.TaxCategoryID,
		IsActive:      true,
	}

	if err := s.repo.Create(ctx, menuItem); err != nil {
		return nil, fmt.Errorf("failed to create menu item: %w", err)
	}

	return menuItem, nil
}

// GetMenuItem retrieves a menu item by ID
func (s *Service) GetMenuItem(ctx context.Context, id uuid.UUID) (*MenuItem, error) {
	return s.repo.GetByID(ctx, id)
}

// ListMenuItems lists menu items for a restaurant with pagination
func (s *Service) ListMenuItems(ctx context.Context, restaurantID uuid.UUID, limit, offset int) ([]*MenuItem, error) {
	return s.repo.ListByRestaurant(ctx, restaurantID, limit, offset)
}

// UpdateMenuItem updates a menu item
func (s *Service) UpdateMenuItem(ctx context.Context, id uuid.UUID, restaurantID uuid.UUID, req *CreateMenuItemRequest) (*MenuItem, error) {
	menuItem, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get menu item: %w", err)
	}

	if menuItem.RestaurantID != restaurantID {
		return nil, fmt.Errorf("menu item does not belong to restaurant")
	}

	menuItem.Name = req.Name
	menuItem.Description = req.Description
	menuItem.Category = req.Category
	menuItem.BasePrice = req.BasePrice
	menuItem.TaxCategoryID = req.TaxCategoryID

	if err := s.repo.Update(ctx, menuItem); err != nil {
		return nil, fmt.Errorf("failed to update menu item: %w", err)
	}

	return menuItem, nil
}

// DeleteMenuItem deletes a menu item (soft delete)
func (s *Service) DeleteMenuItem(ctx context.Context, id uuid.UUID, restaurantID uuid.UUID) error {
	menuItem, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get menu item: %w", err)
	}

	if menuItem.RestaurantID != restaurantID {
		return fmt.Errorf("menu item does not belong to restaurant")
	}

	return s.repo.Delete(ctx, id)
}
