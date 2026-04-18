package promotions

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PromotionType represents the type of promotion
type PromotionType string

const (
	PromotionTypeCoupon   PromotionType = "coupon"
	PromotionTypeDiscount PromotionType = "discount"
	PromotionTypeFreeDelivery PromotionType = "free_delivery"
)

// DiscountType represents how the discount is applied
type DiscountType string

const (
	DiscountTypePercentage DiscountType = "percentage"
	DiscountTypeFixed      DiscountType = "fixed"
)

// Promotion represents a promotion/coupon
type Promotion struct {
	ID              uuid.UUID
	Code            string
	Type            PromotionType
	DiscountType    DiscountType
	DiscountValue   int // Percentage (0-100) or fixed amount in cents
	MinOrderValue   int // Minimum order value in cents to apply promotion
	MaxDiscount     int // Maximum discount amount in cents
	UsageLimit      int // Maximum number of times it can be used (0 = unlimited)
	UsageCount      int // Current usage count
	RegionID        *uuid.UUID
	RestaurantID    *uuid.UUID
	ValidFrom       time.Time
	ValidUntil      time.Time
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Repository defines the interface for promotion data access
type Repository interface {
	Create(ctx context.Context, promotion *Promotion) error
	GetByID(ctx context.Context, id uuid.UUID) (*Promotion, error)
	GetByCode(ctx context.Context, code string) (*Promotion, error)
	List(ctx context.Context, regionID *uuid.UUID, restaurantID *uuid.UUID, activeOnly bool) ([]*Promotion, error)
	Update(ctx context.Context, promotion *Promotion) error
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementUsage(ctx context.Context, id uuid.UUID) error
}

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRepository creates a new promotions repository
func NewRepository(db *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

// Create creates a new promotion
func (r *RepositoryImpl) Create(ctx context.Context, promotion *Promotion) error {
	query := `
		INSERT INTO promotions (id, code, type, discount_type, discount_value, min_order_value, max_discount, usage_limit, usage_count, region_id, restaurant_id, valid_from, valid_until, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := r.db.Exec(ctx, query,
		promotion.ID,
		promotion.Code,
		promotion.Type,
		promotion.DiscountType,
		promotion.DiscountValue,
		promotion.MinOrderValue,
		promotion.MaxDiscount,
		promotion.UsageLimit,
		promotion.UsageCount,
		promotion.RegionID,
		promotion.RestaurantID,
		promotion.ValidFrom,
		promotion.ValidUntil,
		promotion.IsActive,
		promotion.CreatedAt,
		promotion.UpdatedAt,
	)
	return err
}

// GetByID retrieves a promotion by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*Promotion, error) {
	query := `
		SELECT id, code, type, discount_type, discount_value, min_order_value, max_discount, usage_limit, usage_count, region_id, restaurant_id, valid_from, valid_until, is_active, created_at, updated_at
		FROM promotions
		WHERE id = $1 AND deleted_at IS NULL
	`

	var promotion Promotion
	err := r.db.QueryRow(ctx, query, id).Scan(
		&promotion.ID,
		&promotion.Code,
		&promotion.Type,
		&promotion.DiscountType,
		&promotion.DiscountValue,
		&promotion.MinOrderValue,
		&promotion.MaxDiscount,
		&promotion.UsageLimit,
		&promotion.UsageCount,
		&promotion.RegionID,
		&promotion.RestaurantID,
		&promotion.ValidFrom,
		&promotion.ValidUntil,
		&promotion.IsActive,
		&promotion.CreatedAt,
		&promotion.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &promotion, nil
}

// GetByCode retrieves a promotion by code
func (r *RepositoryImpl) GetByCode(ctx context.Context, code string) (*Promotion, error) {
	query := `
		SELECT id, code, type, discount_type, discount_value, min_order_value, max_discount, usage_limit, usage_count, region_id, restaurant_id, valid_from, valid_until, is_active, created_at, updated_at
		FROM promotions
		WHERE code = $1 AND deleted_at IS NULL
	`

	var promotion Promotion
	err := r.db.QueryRow(ctx, query, code).Scan(
		&promotion.ID,
		&promotion.Code,
		&promotion.Type,
		&promotion.DiscountType,
		&promotion.DiscountValue,
		&promotion.MinOrderValue,
		&promotion.MaxDiscount,
		&promotion.UsageLimit,
		&promotion.UsageCount,
		&promotion.RegionID,
		&promotion.RestaurantID,
		&promotion.ValidFrom,
		&promotion.ValidUntil,
		&promotion.IsActive,
		&promotion.CreatedAt,
		&promotion.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &promotion, nil
}

// List lists promotions
func (r *RepositoryImpl) List(ctx context.Context, regionID *uuid.UUID, restaurantID *uuid.UUID, activeOnly bool) ([]*Promotion, error) {
	query := `
		SELECT id, code, type, discount_type, discount_value, min_order_value, max_discount, usage_limit, usage_count, region_id, restaurant_id, valid_from, valid_until, is_active, created_at, updated_at
		FROM promotions
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	argIndex := 1

	if regionID != nil {
		query += " AND region_id = $" + string(rune('0'+argIndex))
		args = append(args, regionID)
		argIndex++
	}

	if restaurantID != nil {
		query += " AND restaurant_id = $" + string(rune('0'+argIndex))
		args = append(args, restaurantID)
		argIndex++
	}

	if activeOnly {
		query += " AND is_active = true"
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var promotions []*Promotion
	for rows.Next() {
		var promotion Promotion
		err := rows.Scan(
			&promotion.ID,
			&promotion.Code,
			&promotion.Type,
			&promotion.DiscountType,
			&promotion.DiscountValue,
			&promotion.MinOrderValue,
			&promotion.MaxDiscount,
			&promotion.UsageLimit,
			&promotion.UsageCount,
			&promotion.RegionID,
			&promotion.RestaurantID,
			&promotion.ValidFrom,
			&promotion.ValidUntil,
			&promotion.IsActive,
			&promotion.CreatedAt,
			&promotion.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		promotions = append(promotions, &promotion)
	}

	return promotions, nil
}

// Update updates a promotion
func (r *RepositoryImpl) Update(ctx context.Context, promotion *Promotion) error {
	query := `
		UPDATE promotions
		SET code = $2, type = $3, discount_type = $4, discount_value = $5, min_order_value = $6, max_discount = $7, usage_limit = $8, region_id = $9, restaurant_id = $10, valid_from = $11, valid_until = $12, is_active = $13, updated_at = $14
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		promotion.ID,
		promotion.Code,
		promotion.Type,
		promotion.DiscountType,
		promotion.DiscountValue,
		promotion.MinOrderValue,
		promotion.MaxDiscount,
		promotion.UsageLimit,
		promotion.RegionID,
		promotion.RestaurantID,
		promotion.ValidFrom,
		promotion.ValidUntil,
		promotion.IsActive,
		promotion.UpdatedAt,
	)
	return err
}

// Delete deletes a promotion (soft delete)
func (r *RepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE promotions
		SET deleted_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	return err
}

// IncrementUsage increments the usage count of a promotion
func (r *RepositoryImpl) IncrementUsage(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE promotions
		SET usage_count = usage_count + 1, updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	return err
}
