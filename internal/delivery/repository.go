package delivery

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for delivery data access
type Repository interface {
	Create(ctx context.Context, delivery *Delivery) error
	GetByID(ctx context.Context, id uuid.UUID) (*Delivery, error)
	ListByCourier(ctx context.Context, courierID uuid.UUID) ([]*Delivery, error)
	ListAvailableCouriers(ctx context.Context) ([]*CourierCandidate, error)
	AssignCourier(ctx context.Context, deliveryID uuid.UUID, courierID uuid.UUID) error
	SetEstimatedDeliveryETA(ctx context.Context, deliveryID uuid.UUID, etaMinutes int) error
	UnassignCourier(ctx context.Context, deliveryID uuid.UUID) error
	UpdateStatus(ctx context.Context, deliveryID uuid.UUID, status string) error
	UpsertCourierLocation(ctx context.Context, courierID uuid.UUID, latitude, longitude float64) error
	ListReassignmentCandidates(ctx context.Context, timeout time.Duration) ([]*Delivery, error)
	ListDeliveriesPastSLA(ctx context.Context, threshold time.Duration) ([]*Delivery, error)
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
		INSERT INTO deliveries (
			id, order_id, courier_id, status,
			pickup_address_street, pickup_address_city, pickup_address_postal_code, pickup_address_country_code,
			delivery_address_street, delivery_address_city, delivery_address_postal_code, delivery_address_country_code,
			pickup_address_latitude, pickup_address_longitude,
			delivery_address_latitude, delivery_address_longitude,
			picked_up_at, delivered_at, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4,
			'', '', '', '',
			'', '', '', '',
			$5, $6, $7, $8,
			$9, $10, NOW() AT TIME ZONE 'UTC', NOW() AT TIME ZONE 'UTC'
		)
	`

	_, err := r.db.Exec(ctx, query,
		delivery.ID,
		delivery.OrderID,
		delivery.CourierID,
		delivery.Status,
		delivery.PickupLocation.Latitude,
		delivery.PickupLocation.Longitude,
		delivery.DeliveryLocation.Latitude,
		delivery.DeliveryLocation.Longitude,
		delivery.PickedUpAt,
		delivery.DeliveredAt,
	)

	return err
}

// GetByID retrieves a delivery by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*Delivery, error) {
	query := `
		SELECT id, order_id, courier_id, status,
			COALESCE(pickup_address_latitude, 0), COALESCE(pickup_address_longitude, 0),
			COALESCE(delivery_address_latitude, 0), COALESCE(delivery_address_longitude, 0),
			estimated_delivery_at, assigned_at, picked_up_at, delivered_at, created_at
		FROM deliveries
		WHERE id = $1
	`

	var delivery Delivery
	err := r.db.QueryRow(ctx, query, id).Scan(
		&delivery.ID,
		&delivery.OrderID,
		&delivery.CourierID,
		&delivery.Status,
		&delivery.PickupLocation.Latitude,
		&delivery.PickupLocation.Longitude,
		&delivery.DeliveryLocation.Latitude,
		&delivery.DeliveryLocation.Longitude,
		&delivery.EstimatedDeliveryAt,
		&delivery.AssignedAt,
		&delivery.PickedUpAt,
		&delivery.DeliveredAt,
		&delivery.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &delivery, nil
}

// ListByCourier lists deliveries for a courier
func (r *RepositoryImpl) ListByCourier(ctx context.Context, courierID uuid.UUID) ([]*Delivery, error) {
	query := `
		SELECT id, order_id, courier_id, status,
			COALESCE(pickup_address_latitude, 0), COALESCE(pickup_address_longitude, 0),
			COALESCE(delivery_address_latitude, 0), COALESCE(delivery_address_longitude, 0),
			estimated_delivery_at, assigned_at, picked_up_at, delivered_at, created_at
		FROM deliveries
		WHERE courier_id = $1
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
			&delivery.PickupLocation.Latitude,
			&delivery.PickupLocation.Longitude,
			&delivery.DeliveryLocation.Latitude,
			&delivery.DeliveryLocation.Longitude,
			&delivery.EstimatedDeliveryAt,
			&delivery.AssignedAt,
			&delivery.PickedUpAt,
			&delivery.DeliveredAt,
			&delivery.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		deliveries = append(deliveries, &delivery)
	}

	return deliveries, nil
}

// ListAvailableCouriers lists couriers that can receive assignments.
func (r *RepositoryImpl) ListAvailableCouriers(ctx context.Context) ([]*CourierCandidate, error) {
	query := `
		SELECT user_id, COALESCE(current_latitude, 0), COALESCE(current_longitude, 0)
		FROM couriers
		WHERE is_available = true
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var couriers []*CourierCandidate
	for rows.Next() {
		candidate := &CourierCandidate{}
		if err := rows.Scan(&candidate.CourierID, &candidate.Location.Latitude, &candidate.Location.Longitude); err != nil {
			return nil, err
		}
		couriers = append(couriers, candidate)
	}

	return couriers, rows.Err()
}

// AssignCourier assigns a courier to a delivery
func (r *RepositoryImpl) AssignCourier(ctx context.Context, deliveryID uuid.UUID, courierID uuid.UUID) error {
	query := `
		UPDATE deliveries
		SET courier_id = $2, status = 'assigned', assigned_at = NOW() AT TIME ZONE 'UTC', updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, deliveryID, courierID)
	return err
}

// SetEstimatedDeliveryETA updates delivery ETA based on route prediction.
func (r *RepositoryImpl) SetEstimatedDeliveryETA(ctx context.Context, deliveryID uuid.UUID, etaMinutes int) error {
	query := `
		UPDATE deliveries
		SET estimated_delivery_at = NOW() AT TIME ZONE 'UTC' + make_interval(mins => $2),
			updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, deliveryID, etaMinutes)
	return err
}

// UnassignCourier removes courier assignment to allow reassignment.
func (r *RepositoryImpl) UnassignCourier(ctx context.Context, deliveryID uuid.UUID) error {
	query := `
		UPDATE deliveries
		SET courier_id = NULL, status = 'unassigned', assigned_at = NULL, updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, deliveryID)
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

// UpsertCourierLocation stores the latest courier GPS location.
func (r *RepositoryImpl) UpsertCourierLocation(ctx context.Context, courierID uuid.UUID, latitude, longitude float64) error {
	query := `
		INSERT INTO couriers (user_id, is_available, current_latitude, current_longitude, current_location_updated_at, created_at, updated_at)
		VALUES ($1, true, $2, $3, NOW() AT TIME ZONE 'UTC', NOW() AT TIME ZONE 'UTC', NOW() AT TIME ZONE 'UTC')
		ON CONFLICT (user_id)
		DO UPDATE SET
			current_latitude = EXCLUDED.current_latitude,
			current_longitude = EXCLUDED.current_longitude,
			current_location_updated_at = EXCLUDED.current_location_updated_at,
			updated_at = NOW() AT TIME ZONE 'UTC'
	`

	if _, err := r.db.Exec(ctx, query, courierID, latitude, longitude); err != nil {
		return err
	}

	locationEventQuery := `
		INSERT INTO courier_location_events (id, courier_user_id, latitude, longitude, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, NOW() AT TIME ZONE 'UTC')
	`
	_, err := r.db.Exec(ctx, locationEventQuery, courierID, latitude, longitude)
	return err
}

// ListReassignmentCandidates returns assigned deliveries that exceeded timeout without pickup.
func (r *RepositoryImpl) ListReassignmentCandidates(ctx context.Context, timeout time.Duration) ([]*Delivery, error) {
	query := `
		SELECT id, order_id, courier_id, status,
			COALESCE(pickup_address_latitude, 0), COALESCE(pickup_address_longitude, 0),
			COALESCE(delivery_address_latitude, 0), COALESCE(delivery_address_longitude, 0),
			estimated_delivery_at, assigned_at, picked_up_at, delivered_at, created_at
		FROM deliveries
		WHERE status = 'assigned'
		  AND picked_up_at IS NULL
		  AND assigned_at IS NOT NULL
		  AND assigned_at < NOW() AT TIME ZONE 'UTC' - ($1::interval)
	`

	rows, err := r.db.Query(ctx, query, timeout.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []*Delivery
	for rows.Next() {
		delivery := &Delivery{}
		if err := rows.Scan(
			&delivery.ID,
			&delivery.OrderID,
			&delivery.CourierID,
			&delivery.Status,
			&delivery.PickupLocation.Latitude,
			&delivery.PickupLocation.Longitude,
			&delivery.DeliveryLocation.Latitude,
			&delivery.DeliveryLocation.Longitude,
			&delivery.EstimatedDeliveryAt,
			&delivery.AssignedAt,
			&delivery.PickedUpAt,
			&delivery.DeliveredAt,
			&delivery.CreatedAt,
		); err != nil {
			return nil, err
		}
		deliveries = append(deliveries, delivery)
	}

	return deliveries, rows.Err()
}

// ListDeliveriesPastSLA returns deliveries that breached SLA threshold.
func (r *RepositoryImpl) ListDeliveriesPastSLA(ctx context.Context, threshold time.Duration) ([]*Delivery, error) {
	query := `
		SELECT id, order_id, courier_id, status,
			COALESCE(pickup_address_latitude, 0), COALESCE(pickup_address_longitude, 0),
			COALESCE(delivery_address_latitude, 0), COALESCE(delivery_address_longitude, 0),
			estimated_delivery_at, assigned_at, picked_up_at, delivered_at, created_at
		FROM deliveries
		WHERE status IN ('assigned', 'picked_up')
		  AND delivered_at IS NULL
		  AND created_at < NOW() AT TIME ZONE 'UTC' - ($1::interval)
	`

	rows, err := r.db.Query(ctx, query, threshold.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []*Delivery
	for rows.Next() {
		delivery := &Delivery{}
		if err := rows.Scan(
			&delivery.ID,
			&delivery.OrderID,
			&delivery.CourierID,
			&delivery.Status,
			&delivery.PickupLocation.Latitude,
			&delivery.PickupLocation.Longitude,
			&delivery.DeliveryLocation.Latitude,
			&delivery.DeliveryLocation.Longitude,
			&delivery.EstimatedDeliveryAt,
			&delivery.AssignedAt,
			&delivery.PickedUpAt,
			&delivery.DeliveredAt,
			&delivery.CreatedAt,
		); err != nil {
			return nil, err
		}
		deliveries = append(deliveries, delivery)
	}

	return deliveries, rows.Err()
}

// MarkPickedUp marks a delivery as picked up
func (r *RepositoryImpl) MarkPickedUp(ctx context.Context, deliveryID uuid.UUID) error {
	query := `
		UPDATE deliveries
		SET status = 'picked_up', picked_up_at = NOW() AT TIME ZONE 'UTC', updated_at = NOW() AT TIME ZONE 'UTC'
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
