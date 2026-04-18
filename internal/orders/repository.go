package orders

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for order data access
type Repository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*Order, error)
	ListByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]*Order, error)
	UpdateWorkflowState(ctx context.Context, id uuid.UUID, state string) error
}

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRepository creates a new orders repository
func NewRepository(db *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

// Create creates a new order
func (r *RepositoryImpl) Create(ctx context.Context, order *Order) error {
	query := `
		INSERT INTO orders (id, customer_id, restaurant_id, region_id, delivery_address_id, workflow_state, subtotal, delivery_fee, tax, total, currency_code, scheduled_for, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.db.Exec(ctx, query,
		order.ID,
		order.CustomerID,
		order.RestaurantID,
		order.RegionID,
		order.DeliveryAddressID,
		order.WorkflowState,
		order.Subtotal,
		order.DeliveryFee,
		order.Tax,
		order.Total,
		order.CurrencyCode,
		order.ScheduledFor,
		order.CreatedAt,
		order.CreatedAt,
	)

	return err
}

// GetByID retrieves an order by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*Order, error) {
	query := `
		SELECT id, customer_id, restaurant_id, region_id, delivery_address_id, workflow_state, subtotal, delivery_fee, tax, total, currency_code, scheduled_for, created_at
		FROM orders
		WHERE id = $1 AND deleted_at IS NULL
	`

	var order Order
	err := r.db.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.RestaurantID,
		&order.RegionID,
		&order.DeliveryAddressID,
		&order.WorkflowState,
		&order.Subtotal,
		&order.DeliveryFee,
		&order.Tax,
		&order.Total,
		&order.CurrencyCode,
		&order.ScheduledFor,
		&order.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &order, nil
}

// ListByCustomer lists orders for a customer
func (r *RepositoryImpl) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*Order, error) {
	query := `
		SELECT id, customer_id, restaurant_id, region_id, delivery_address_id, workflow_state, subtotal, delivery_fee, tax, total, currency_code, scheduled_for, created_at
		FROM orders
		WHERE customer_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var order Order
		err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.RestaurantID,
			&order.RegionID,
			&order.DeliveryAddressID,
			&order.WorkflowState,
			&order.Subtotal,
			&order.DeliveryFee,
			&order.Tax,
			&order.Total,
			&order.CurrencyCode,
			&order.ScheduledFor,
			&order.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

// ListByRestaurant lists orders for a restaurant
func (r *RepositoryImpl) ListByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]*Order, error) {
	query := `
		SELECT id, customer_id, restaurant_id, region_id, delivery_address_id, workflow_state, subtotal, delivery_fee, tax, total, currency_code, scheduled_for, created_at
		FROM orders
		WHERE restaurant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var order Order
		err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.RestaurantID,
			&order.RegionID,
			&order.DeliveryAddressID,
			&order.WorkflowState,
			&order.Subtotal,
			&order.DeliveryFee,
			&order.Tax,
			&order.Total,
			&order.CurrencyCode,
			&order.ScheduledFor,
			&order.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

// UpdateWorkflowState updates an order's workflow state
func (r *RepositoryImpl) UpdateWorkflowState(ctx context.Context, id uuid.UUID, state string) error {
	query := `
		UPDATE orders
		SET workflow_state = $2, updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, state)
	return err
}
