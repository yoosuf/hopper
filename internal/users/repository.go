package users

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Role         string
	FirstName    string
	LastName     string
	IsSuspended  bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Repository defines the interface for user data access
type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	Update(ctx context.Context, user *User) error
	CreateAddress(ctx context.Context, address *Address) error
	ListAddresses(ctx context.Context, userID uuid.UUID) ([]*Address, error)
	GetAddress(ctx context.Context, id uuid.UUID) (*Address, error)
	UpdateAddress(ctx context.Context, address *Address) error
	DeleteAddress(ctx context.Context, id uuid.UUID) error
}

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRepository creates a new users repository
func NewRepository(db *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

// GetByID retrieves a user by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, email, password_hash, role, first_name, last_name, is_suspended, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.FirstName,
		&user.LastName,
		&user.IsSuspended,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Update updates a user
func (r *RepositoryImpl) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET first_name = $2, last_name = $3, updated_at = $4
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, user.ID, user.FirstName, user.LastName, time.Now().UTC())
	return err
}

// CreateAddress creates a new address
func (r *RepositoryImpl) CreateAddress(ctx context.Context, address *Address) error {
	query := `
		INSERT INTO user_addresses (id, user_id, street_address, city, state_or_province, postal_code, country_code, latitude, longitude, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	now := time.Now().UTC()
	_, err := r.db.Exec(ctx, query,
		address.ID,
		address.UserID,
		address.StreetAddress,
		address.City,
		address.State,
		address.PostalCode,
		address.CountryCode,
		address.Latitude,
		address.Longitude,
		address.IsDefault,
		now,
		now,
	)

	return err
}

// ListAddresses lists all addresses for a user
func (r *RepositoryImpl) ListAddresses(ctx context.Context, userID uuid.UUID) ([]*Address, error) {
	query := `
		SELECT id, user_id, street_address, city, state_or_province, postal_code, country_code, latitude, longitude, is_default
		FROM user_addresses
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY is_default DESC, created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []*Address
	for rows.Next() {
		var addr Address
		err := rows.Scan(
			&addr.ID,
			&addr.UserID,
			&addr.StreetAddress,
			&addr.City,
			&addr.State,
			&addr.PostalCode,
			&addr.CountryCode,
			&addr.Latitude,
			&addr.Longitude,
			&addr.IsDefault,
		)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, &addr)
	}

	return addresses, nil
}

// GetAddress retrieves an address by ID
func (r *RepositoryImpl) GetAddress(ctx context.Context, id uuid.UUID) (*Address, error) {
	query := `
		SELECT id, user_id, street_address, city, state_or_province, postal_code, country_code, latitude, longitude, is_default
		FROM user_addresses
		WHERE id = $1 AND deleted_at IS NULL
	`

	var addr Address
	err := r.db.QueryRow(ctx, query, id).Scan(
		&addr.ID,
		&addr.UserID,
		&addr.StreetAddress,
		&addr.City,
		&addr.State,
		&addr.PostalCode,
		&addr.CountryCode,
		&addr.Latitude,
		&addr.Longitude,
		&addr.IsDefault,
	)

	if err != nil {
		return nil, err
	}

	return &addr, nil
}

// UpdateAddress updates an address
func (r *RepositoryImpl) UpdateAddress(ctx context.Context, address *Address) error {
	query := `
		UPDATE user_addresses
		SET street_address = $2, city = $3, state_or_province = $4, postal_code = $5, country_code = $6, latitude = $7, longitude = $8, is_default = $9, updated_at = $10
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		address.ID,
		address.StreetAddress,
		address.City,
		address.State,
		address.PostalCode,
		address.CountryCode,
		address.Latitude,
		address.Longitude,
		address.IsDefault,
		time.Now().UTC(),
	)

	return err
}

// DeleteAddress soft deletes an address
func (r *RepositoryImpl) DeleteAddress(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE user_addresses
		SET deleted_at = $1
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, time.Now().UTC(), id)
	return err
}
