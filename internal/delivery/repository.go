package delivery

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for delivery data access
type Repository interface {
	Create(ctx context.Context, delivery *Delivery) error
	GetByID(ctx context.Context, id uuid.UUID) (*Delivery, error)
	ListByCourier(ctx context.Context, courierID uuid.UUID) ([]*Delivery, error)
	AssignCourier(ctx context.Context, deliveryID uuid.UUID, courierID uuid.UUID) error
	UpdateStatus(ctx context.Context, deliveryID uuid.UUID, status string) error
	MarkPickedUp(ctx context.Context, deliveryID uuid.UUID) error
	MarkDelivered(ctx context.Context, deliveryID uuid.UUID) error
}

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRepository creates a new delivery repository
func NewRepository(db *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

// Create creates a new delivery
func (r *RepositoryImpl) Create(ctx context.Context, delivery *Delivery) error {
	query := `
		INSERT INTO deliveries (id, order_id, courier_id, status, picked_up_at, delivered_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW() AT TIME ZONE 'UTC', NOW() AT TIME ZONE 'UTC')
	`

	_, err := r.db.Exec(ctx, query,
		delivery.ID,
		delivery.OrderID,
		delivery.CourierID,
		delivery.Status,
		delivery.PickedUpAt,
		delivery.DeliveredAt,
	)

	return err
}

// GetByID retrieves a delivery by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*Delivery, error) {
	query := `
		SELECT id, order_id, courier_id, status, picked_up_at, delivered_at
		FROM deliveries
		WHERE id = $1 AND deleted_at IS NULL
	`

	var delivery Delivery
	err := r.db.QueryRow(ctx, query, id).Scan(
		&delivery.ID,
		&delivery.OrderID,
		&delivery.CourierID,
		&delivery.Status,
		&delivery.PickedUpAt,
		&delivery.DeliveredAt,
	)

	if err != nil {
		return nil, err
	}

	return &delivery, nil
}

// ListByCourier lists deliveries for a courier
func (r *RepositoryImpl) ListByCourier(ctx context.Context, courierID uuid.UUID) ([]*Delivery, error) {
	query := `
		SELECT id, order_id, courier_id, status, picked_up_at, delivered_at
		FROM deliveries
		WHERE courier_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, courierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []*Delivery
	for rows.Next() {
		var delivery Delivery
		err := rows.Scan(
			&delivery.ID,
			&delivery.OrderID,
			&delivery.CourierID,
			&delivery.Status,
			&delivery.PickedUpAt,
			&delivery.DeliveredAt,
		)
		if err != nil {
			return nil, err
		}
		deliveries = append(deliveries, &delivery)
	}

	return deliveries, nil
}

// AssignCourier assigns a courier to a delivery
func (r *RepositoryImpl) AssignCourier(ctx context.Context, deliveryID uuid.UUID, courierID uuid.UUID) error {
	query := `
		UPDATE deliveries
		SET courier_id = $2, status = 'assigned', updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, deliveryID, courierID)
	return err
}

// UpdateStatus updates a delivery's status
func (r *RepositoryImpl) UpdateStatus(ctx context.Context, deliveryID uuid.UUID, status string) error {
	query := `
		UPDATE deliveries
		SET status = $2, updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, deliveryID, status)
	return err
}

// MarkPickedUp marks a delivery as picked up
func (r *RepositoryImpl) MarkPickedUp(ctx context.Context, deliveryID uuid.UUID) error {
	query := `
		UPDATE deliveries
		SET status = 'in_transit', picked_up_at = NOW() AT TIME ZONE 'UTC', updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, deliveryID)
	return err
}

// MarkDelivered marks a delivery as delivered
func (r *RepositoryImpl) MarkDelivered(ctx context.Context, deliveryID uuid.UUID) error {
	query := `
		UPDATE deliveries
		SET status = 'delivered', delivered_at = NOW() AT TIME ZONE 'UTC', updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, deliveryID)
	return err
}
