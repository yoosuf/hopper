package admin

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for admin data access
type Repository interface {
	ApproveRestaurant(ctx context.Context, restaurantID uuid.UUID) error
	RejectRestaurant(ctx context.Context, restaurantID uuid.UUID) error
	ListPendingRestaurants(ctx context.Context) ([]uuid.UUID, error)
	GetSystemStats(ctx context.Context) (*SystemStats, error)
}

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRepository creates a new admin repository
func NewRepository(db *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

// ApproveRestaurant approves a restaurant
func (r *RepositoryImpl) ApproveRestaurant(ctx context.Context, restaurantID uuid.UUID) error {
	query := `
		UPDATE restaurants
		SET is_approved = true, approved_at = NOW() AT TIME ZONE 'UTC', updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, restaurantID)
	return err
}

// RejectRestaurant rejects a restaurant
func (r *RepositoryImpl) RejectRestaurant(ctx context.Context, restaurantID uuid.UUID) error {
	query := `
		UPDATE restaurants
		SET is_active = false, updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, restaurantID)
	return err
}

// ListPendingRestaurants lists restaurants pending approval
func (r *RepositoryImpl) ListPendingRestaurants(ctx context.Context) ([]uuid.UUID, error) {
	query := `
		SELECT id
		FROM restaurants
		WHERE is_approved = false AND is_active = true AND deleted_at IS NULL
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var restaurantIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		restaurantIDs = append(restaurantIDs, id)
	}

	return restaurantIDs, nil
}

// GetSystemStats retrieves system statistics
func (r *RepositoryImpl) GetSystemStats(ctx context.Context) (*SystemStats, error) {
	stats := &SystemStats{}

	query := `
		SELECT
			(SELECT COUNT(*) FROM users WHERE deleted_at IS NULL) as total_users,
			(SELECT COUNT(*) FROM restaurants WHERE deleted_at IS NULL) as total_restaurants,
			(SELECT COUNT(*) FROM orders WHERE deleted_at IS NULL) as total_orders,
			(SELECT COUNT(*) FROM deliveries WHERE status = 'in_transit' AND deleted_at IS NULL) as active_deliveries
	`

	err := r.db.QueryRow(ctx, query).Scan(
		&stats.TotalUsers,
		&stats.TotalRestaurants,
		&stats.TotalOrders,
		&stats.ActiveDeliveries,
	)

	if err != nil {
		return nil, err
	}

	return stats, nil
}
