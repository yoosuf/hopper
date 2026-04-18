package promotions

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles promotion operations
type Service struct {
	repo Repository
}

// New creates a new promotions service
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreatePromotionRequest represents a create promotion request
type CreatePromotionRequest struct {
	Code          string        `json:"code" validate:"required,min=3,max=50"`
	Type          PromotionType `json:"type" validate:"required,oneof=coupon discount free_delivery"`
	DiscountType  DiscountType  `json:"discount_type" validate:"required,oneof=percentage fixed"`
	DiscountValue int           `json:"discount_value" validate:"required,min=1"`
	MinOrderValue int           `json:"min_order_value" validate:"min=0"`
	MaxDiscount   int           `json:"max_discount" validate:"min=0"`
	UsageLimit    int           `json:"usage_limit" validate:"min=0"`
	RegionID      *uuid.UUID    `json:"region_id"`
	RestaurantID  *uuid.UUID    `json:"restaurant_id"`
	ValidFrom     time.Time     `json:"valid_from" validate:"required"`
	ValidUntil    time.Time     `json:"valid_until" validate:"required"`
}

// UpdatePromotionRequest represents an update promotion request
type UpdatePromotionRequest struct {
	Code          string        `json:"code" validate:"min=3,max=50"`
	Type          PromotionType `json:"type" validate:"oneof=coupon discount free_delivery"`
	DiscountType  DiscountType  `json:"discount_type" validate:"oneof=percentage fixed"`
	DiscountValue int           `json:"discount_value" validate:"min=1"`
	MinOrderValue int           `json:"min_order_value" validate:"min=0"`
	MaxDiscount   int           `json:"max_discount" validate:"min=0"`
	UsageLimit    int           `json:"usage_limit" validate:"min=0"`
	RegionID      *uuid.UUID    `json:"region_id"`
	RestaurantID  *uuid.UUID    `json:"restaurant_id"`
	ValidFrom     time.Time     `json:"valid_from"`
	ValidUntil    time.Time     `json:"valid_until"`
	IsActive      *bool         `json:"is_active"`
}

// PromotionResponse represents a promotion response
type PromotionResponse struct {
	ID            uuid.UUID     `json:"id"`
	Code          string        `json:"code"`
	Type          PromotionType `json:"type"`
	DiscountType  DiscountType  `json:"discount_type"`
	DiscountValue int           `json:"discount_value"`
	MinOrderValue int           `json:"min_order_value"`
	MaxDiscount   int           `json:"max_discount"`
	UsageLimit    int           `json:"usage_limit"`
	UsageCount    int           `json:"usage_count"`
	RegionID      *uuid.UUID    `json:"region_id"`
	RestaurantID  *uuid.UUID    `json:"restaurant_id"`
	ValidFrom     time.Time     `json:"valid_from"`
	ValidUntil    time.Time     `json:"valid_until"`
	IsActive      bool          `json:"is_active"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// CreatePromotion creates a new promotion
func (s *Service) CreatePromotion(ctx context.Context, req *CreatePromotionRequest) (*PromotionResponse, error) {
	// Check if code already exists
	existing, err := s.repo.GetByCode(ctx, req.Code)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("promotion code already exists")
	}

	promotion := &Promotion{
		ID:            uuid.New(),
		Code:          req.Code,
		Type:          req.Type,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		MinOrderValue: req.MinOrderValue,
		MaxDiscount:   req.MaxDiscount,
		UsageLimit:    req.UsageLimit,
		UsageCount:    0,
		RegionID:      req.RegionID,
		RestaurantID:  req.RestaurantID,
		ValidFrom:     req.ValidFrom,
		ValidUntil:    req.ValidUntil,
		IsActive:      true,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, promotion); err != nil {
		return nil, fmt.Errorf("failed to create promotion: %w", err)
	}

	return &PromotionResponse{
		ID:            promotion.ID,
		Code:          promotion.Code,
		Type:          promotion.Type,
		DiscountType:  promotion.DiscountType,
		DiscountValue: promotion.DiscountValue,
		MinOrderValue: promotion.MinOrderValue,
		MaxDiscount:   promotion.MaxDiscount,
		UsageLimit:    promotion.UsageLimit,
		UsageCount:    promotion.UsageCount,
		RegionID:      promotion.RegionID,
		RestaurantID:  promotion.RestaurantID,
		ValidFrom:     promotion.ValidFrom,
		ValidUntil:    promotion.ValidUntil,
		IsActive:      promotion.IsActive,
		CreatedAt:     promotion.CreatedAt,
		UpdatedAt:     promotion.UpdatedAt,
	}, nil
}

// GetPromotion retrieves a promotion by ID
func (s *Service) GetPromotion(ctx context.Context, id uuid.UUID) (*PromotionResponse, error) {
	promotion, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get promotion: %w", err)
	}

	return &PromotionResponse{
		ID:            promotion.ID,
		Code:          promotion.Code,
		Type:          promotion.Type,
		DiscountType:  promotion.DiscountType,
		DiscountValue: promotion.DiscountValue,
		MinOrderValue: promotion.MinOrderValue,
		MaxDiscount:   promotion.MaxDiscount,
		UsageLimit:    promotion.UsageLimit,
		UsageCount:    promotion.UsageCount,
		RegionID:      promotion.RegionID,
		RestaurantID:  promotion.RestaurantID,
		ValidFrom:     promotion.ValidFrom,
		ValidUntil:    promotion.ValidUntil,
		IsActive:      promotion.IsActive,
		CreatedAt:     promotion.CreatedAt,
		UpdatedAt:     promotion.UpdatedAt,
	}, nil
}

// GetPromotionByCode retrieves a promotion by code
func (s *Service) GetPromotionByCode(ctx context.Context, code string) (*PromotionResponse, error) {
	promotion, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get promotion: %w", err)
	}

	return &PromotionResponse{
		ID:            promotion.ID,
		Code:          promotion.Code,
		Type:          promotion.Type,
		DiscountType:  promotion.DiscountType,
		DiscountValue: promotion.DiscountValue,
		MinOrderValue: promotion.MinOrderValue,
		MaxDiscount:   promotion.MaxDiscount,
		UsageLimit:    promotion.UsageLimit,
		UsageCount:    promotion.UsageCount,
		RegionID:      promotion.RegionID,
		RestaurantID:  promotion.RestaurantID,
		ValidFrom:     promotion.ValidFrom,
		ValidUntil:    promotion.ValidUntil,
		IsActive:      promotion.IsActive,
		CreatedAt:     promotion.CreatedAt,
		UpdatedAt:     promotion.UpdatedAt,
	}, nil
}

// ListPromotions lists promotions
func (s *Service) ListPromotions(ctx context.Context, regionID *uuid.UUID, restaurantID *uuid.UUID, activeOnly bool) ([]*PromotionResponse, error) {
	promotions, err := s.repo.List(ctx, regionID, restaurantID, activeOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to list promotions: %w", err)
	}

	var responses []*PromotionResponse
	for _, promotion := range promotions {
		responses = append(responses, &PromotionResponse{
			ID:            promotion.ID,
			Code:          promotion.Code,
			Type:          promotion.Type,
			DiscountType:  promotion.DiscountType,
			DiscountValue: promotion.DiscountValue,
			MinOrderValue: promotion.MinOrderValue,
			MaxDiscount:   promotion.MaxDiscount,
			UsageLimit:    promotion.UsageLimit,
			UsageCount:    promotion.UsageCount,
			RegionID:      promotion.RegionID,
			RestaurantID:  promotion.RestaurantID,
			ValidFrom:     promotion.ValidFrom,
			ValidUntil:    promotion.ValidUntil,
			IsActive:      promotion.IsActive,
			CreatedAt:     promotion.CreatedAt,
			UpdatedAt:     promotion.UpdatedAt,
		})
	}

	return responses, nil
}

// UpdatePromotion updates a promotion
func (s *Service) UpdatePromotion(ctx context.Context, id uuid.UUID, req *UpdatePromotionRequest) (*PromotionResponse, error) {
	promotion, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get promotion: %w", err)
	}

	if req.Code != "" {
		promotion.Code = req.Code
	}
	if req.Type != "" {
		promotion.Type = req.Type
	}
	if req.DiscountType != "" {
		promotion.DiscountType = req.DiscountType
	}
	if req.DiscountValue > 0 {
		promotion.DiscountValue = req.DiscountValue
	}
	if req.MinOrderValue >= 0 {
		promotion.MinOrderValue = req.MinOrderValue
	}
	if req.MaxDiscount >= 0 {
		promotion.MaxDiscount = req.MaxDiscount
	}
	if req.UsageLimit >= 0 {
		promotion.UsageLimit = req.UsageLimit
	}
	if !req.ValidFrom.IsZero() {
		promotion.ValidFrom = req.ValidFrom
	}
	if !req.ValidUntil.IsZero() {
		promotion.ValidUntil = req.ValidUntil
	}
	if req.IsActive != nil {
		promotion.IsActive = *req.IsActive
	}
	promotion.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, promotion); err != nil {
		return nil, fmt.Errorf("failed to update promotion: %w", err)
	}

	return &PromotionResponse{
		ID:            promotion.ID,
		Code:          promotion.Code,
		Type:          promotion.Type,
		DiscountType:  promotion.DiscountType,
		DiscountValue: promotion.DiscountValue,
		MinOrderValue: promotion.MinOrderValue,
		MaxDiscount:   promotion.MaxDiscount,
		UsageLimit:    promotion.UsageLimit,
		UsageCount:    promotion.UsageCount,
		RegionID:      promotion.RegionID,
		RestaurantID:  promotion.RestaurantID,
		ValidFrom:     promotion.ValidFrom,
		ValidUntil:    promotion.ValidUntil,
		IsActive:      promotion.IsActive,
		CreatedAt:     promotion.CreatedAt,
		UpdatedAt:     promotion.UpdatedAt,
	}, nil
}

// DeletePromotion deletes a promotion
func (s *Service) DeletePromotion(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ValidatePromotion validates if a promotion can be applied
func (s *Service) ValidatePromotion(ctx context.Context, code string, orderValue int, regionID *uuid.UUID, restaurantID *uuid.UUID) (*Promotion, error) {
	promotion, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("promotion not found")
	}

	now := time.Now().UTC()

	// Check if promotion is active
	if !promotion.IsActive {
		return nil, fmt.Errorf("promotion is not active")
	}

	// Check if promotion is within valid date range
	if now.Before(promotion.ValidFrom) || now.After(promotion.ValidUntil) {
		return nil, fmt.Errorf("promotion is not valid at this time")
	}

	// Check usage limit
	if promotion.UsageLimit > 0 && promotion.UsageCount >= promotion.UsageLimit {
		return nil, fmt.Errorf("promotion has reached its usage limit")
	}

	// Check minimum order value
	if orderValue < promotion.MinOrderValue {
		return nil, fmt.Errorf("order value is below minimum required")
	}

	// Check region scope
	if promotion.RegionID != nil && regionID != nil && *promotion.RegionID != *regionID {
		return nil, fmt.Errorf("promotion is not valid for this region")
	}

	// Check restaurant scope
	if promotion.RestaurantID != nil && restaurantID != nil && *promotion.RestaurantID != *restaurantID {
		return nil, fmt.Errorf("promotion is not valid for this restaurant")
	}

	return promotion, nil
}

// ApplyPromotion applies a promotion and returns the discount amount
func (s *Service) ApplyPromotion(ctx context.Context, promotion *Promotion, orderValue int) int {
	var discount int

	switch promotion.DiscountType {
	case DiscountTypePercentage:
		discount = orderValue * promotion.DiscountValue / 100
	case DiscountTypeFixed:
		discount = promotion.DiscountValue
	}

	// Apply max discount limit
	if promotion.MaxDiscount > 0 && discount > promotion.MaxDiscount {
		discount = promotion.MaxDiscount
	}

	// Ensure discount doesn't exceed order value
	if discount > orderValue {
		discount = orderValue
	}

	return discount
}

// UsePromotion marks a promotion as used
func (s *Service) UsePromotion(ctx context.Context, id uuid.UUID) error {
	return s.repo.IncrementUsage(ctx, id)
}
