package restaurants

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for restaurant data access
type Repository interface {
	Create(ctx context.Context, restaurant *Restaurant) error
	GetByID(ctx context.Context, id uuid.UUID) (*Restaurant, error)
	ListByRegion(ctx context.Context, regionID uuid.UUID) ([]*Restaurant, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*Restaurant, error)
	Update(ctx context.Context, restaurant *Restaurant) error
	DeleteHours(ctx context.Context, restaurantID uuid.UUID) error
	CreateHour(ctx context.Context, hour *RestaurantHour) error
	ListHours(ctx context.Context, restaurantID uuid.UUID) ([]*RestaurantHour, error)
}

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRepository creates a new restaurants repository
func NewRepository(db *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

// Create creates a new restaurant
func (r *RepositoryImpl) Create(ctx context.Context, restaurant *Restaurant) error {
	query := `
		INSERT INTO restaurants (id, owner_id, name, description, cuisine_type, street_address, city, state_or_province, postal_code, country_code, latitude, longitude, phone, email, region_id, currency_code, timezone, is_active, is_approved, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	`

	now := time.Now().UTC()
	_, err := r.db.Exec(ctx, query,
		restaurant.ID,
		restaurant.OwnerID,
		restaurant.Name,
		restaurant.Description,
		restaurant.CuisineType,
		restaurant.StreetAddress,
		restaurant.City,
		restaurant.State,
		restaurant.PostalCode,
		restaurant.CountryCode,
		restaurant.Latitude,
		restaurant.Longitude,
		restaurant.Phone,
		restaurant.Email,
		restaurant.RegionID,
		restaurant.CurrencyCode,
		restaurant.Timezone,
		restaurant.IsActive,
		restaurant.IsApproved,
		now,
		now,
	)

	return err
}

// GetByID retrieves a restaurant by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*Restaurant, error) {
	query := `
		SELECT id, owner_id, name, description, cuisine_type, street_address, city, state_or_province, postal_code, country_code, latitude, longitude, phone, email, region_id, currency_code, timezone, is_active, is_approved, approved_at
		FROM restaurants
		WHERE id = $1 AND deleted_at IS NULL
	`

	var restaurant Restaurant
	err := r.db.QueryRow(ctx, query, id).Scan(
		&restaurant.ID,
		&restaurant.OwnerID,
		&restaurant.Name,
		&restaurant.Description,
		&restaurant.CuisineType,
		&restaurant.StreetAddress,
		&restaurant.City,
		&restaurant.State,
		&restaurant.PostalCode,
		&restaurant.CountryCode,
		&restaurant.Latitude,
		&restaurant.Longitude,
		&restaurant.Phone,
		&restaurant.Email,
		&restaurant.RegionID,
		&restaurant.CurrencyCode,
		&restaurant.Timezone,
		&restaurant.IsActive,
		&restaurant.IsApproved,
		&restaurant.ApprovedAt,
	)

	if err != nil {
		return nil, err
	}

	return &restaurant, nil
}

// ListByRegion lists restaurants in a region
func (r *RepositoryImpl) ListByRegion(ctx context.Context, regionID uuid.UUID) ([]*Restaurant, error) {
	query := `
		SELECT id, owner_id, name, description, cuisine_type, street_address, city, state_or_province, postal_code, country_code, latitude, longitude, phone, email, region_id, currency_code, timezone, is_active, is_approved, approved_at
		FROM restaurants
		WHERE region_id = $1 AND is_active = true AND is_approved = true AND deleted_at IS NULL
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query, regionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var restaurants []*Restaurant
	for rows.Next() {
		var restaurant Restaurant
		err := rows.Scan(
			&restaurant.ID,
			&restaurant.OwnerID,
			&restaurant.Name,
			&restaurant.Description,
			&restaurant.CuisineType,
			&restaurant.StreetAddress,
			&restaurant.City,
			&restaurant.State,
			&restaurant.PostalCode,
			&restaurant.CountryCode,
			&restaurant.Latitude,
			&restaurant.Longitude,
			&restaurant.Phone,
			&restaurant.Email,
			&restaurant.RegionID,
			&restaurant.CurrencyCode,
			&restaurant.Timezone,
			&restaurant.IsActive,
			&restaurant.IsApproved,
			&restaurant.ApprovedAt,
		)
		if err != nil {
			return nil, err
		}
		restaurants = append(restaurants, &restaurant)
	}

	return restaurants, nil
}

// ListByOwner lists restaurants owned by a user
func (r *RepositoryImpl) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*Restaurant, error) {
	query := `
		SELECT id, owner_id, name, description, cuisine_type, street_address, city, state_or_province, postal_code, country_code, latitude, longitude, phone, email, region_id, currency_code, timezone, is_active, is_approved, approved_at
		FROM restaurants
		WHERE owner_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var restaurants []*Restaurant
	for rows.Next() {
		var restaurant Restaurant
		err := rows.Scan(
			&restaurant.ID,
			&restaurant.OwnerID,
			&restaurant.Name,
			&restaurant.Description,
			&restaurant.CuisineType,
			&restaurant.StreetAddress,
			&restaurant.City,
			&restaurant.State,
			&restaurant.PostalCode,
			&restaurant.CountryCode,
			&restaurant.Latitude,
			&restaurant.Longitude,
			&restaurant.Phone,
			&restaurant.Email,
			&restaurant.RegionID,
			&restaurant.CurrencyCode,
			&restaurant.Timezone,
			&restaurant.IsActive,
			&restaurant.IsApproved,
			&restaurant.ApprovedAt,
		)
		if err != nil {
			return nil, err
		}
		restaurants = append(restaurants, &restaurant)
	}

	return restaurants, nil
}

// Update updates a restaurant
func (r *RepositoryImpl) Update(ctx context.Context, restaurant *Restaurant) error {
	query := `
		UPDATE restaurants
		SET name = $2, description = $3, cuisine_type = $4, street_address = $5, city = $6, state_or_province = $7, postal_code = $8, country_code = $9, latitude = $10, longitude = $11, phone = $12, email = $13, timezone = $14, updated_at = $15
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		restaurant.ID,
		restaurant.Name,
		restaurant.Description,
		restaurant.CuisineType,
		restaurant.StreetAddress,
		restaurant.City,
		restaurant.State,
		restaurant.PostalCode,
		restaurant.CountryCode,
		restaurant.Latitude,
		restaurant.Longitude,
		restaurant.Phone,
		restaurant.Email,
		restaurant.Timezone,
		time.Now().UTC(),
	)

	return err
}

// DeleteHours deletes all hours for a restaurant
func (r *RepositoryImpl) DeleteHours(ctx context.Context, restaurantID uuid.UUID) error {
	query := `DELETE FROM restaurant_hours WHERE restaurant_id = $1`
	_, err := r.db.Exec(ctx, query, restaurantID)
	return err
}

// CreateHour creates a restaurant hour
func (r *RepositoryImpl) CreateHour(ctx context.Context, hour *RestaurantHour) error {
	query := `
		INSERT INTO restaurant_hours (id, restaurant_id, day_of_week, open_time, close_time, is_closed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now().UTC()
	_, err := r.db.Exec(ctx, query,
		uuid.New(),
		hour.RestaurantID,
		hour.DayOfWeek,
		hour.OpenTime,
		hour.CloseTime,
		hour.IsClosed,
		now,
		now,
	)

	return err
}

// ListHours lists restaurant hours
func (r *RepositoryImpl) ListHours(ctx context.Context, restaurantID uuid.UUID) ([]*RestaurantHour, error) {
	query := `
		SELECT id, restaurant_id, day_of_week, open_time, close_time, is_closed
		FROM restaurant_hours
		WHERE restaurant_id = $1
		ORDER BY day_of_week
	`

	rows, err := r.db.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hours []*RestaurantHour
	for rows.Next() {
		var hour RestaurantHour
		err := rows.Scan(
			&hour.ID,
			&hour.RestaurantID,
			&hour.DayOfWeek,
			&hour.OpenTime,
			&hour.CloseTime,
			&hour.IsClosed,
		)
		if err != nil {
			return nil, err
		}
		hours = append(hours, &hour)
	}

	return hours, nil
}
