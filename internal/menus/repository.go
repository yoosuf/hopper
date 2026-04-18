package menus

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for menu data access
type Repository interface {
	Create(ctx context.Context, menuItem *MenuItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*MenuItem, error)
	ListByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]*MenuItem, error)
	Update(ctx context.Context, menuItem *MenuItem) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRepository creates a new menus repository
func NewRepository(db *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

// Create creates a new menu item
func (r *RepositoryImpl) Create(ctx context.Context, menuItem *MenuItem) error {
	query := `
		INSERT INTO menu_items (id, restaurant_id, name, description, category, base_price, tax_category_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now().UTC()
	_, err := r.db.Exec(ctx, query,
		menuItem.ID,
		menuItem.RestaurantID,
		menuItem.Name,
		menuItem.Description,
		menuItem.Category,
		menuItem.BasePrice,
		menuItem.TaxCategoryID,
		menuItem.IsActive,
		now,
		now,
	)

	return err
}

// GetByID retrieves a menu item by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*MenuItem, error) {
	query := `
		SELECT id, restaurant_id, name, description, category, base_price, tax_category_id, is_active
		FROM menu_items
		WHERE id = $1 AND deleted_at IS NULL
	`

	var menuItem MenuItem
	err := r.db.QueryRow(ctx, query, id).Scan(
		&menuItem.ID,
		&menuItem.RestaurantID,
		&menuItem.Name,
		&menuItem.Description,
		&menuItem.Category,
		&menuItem.BasePrice,
		&menuItem.TaxCategoryID,
		&menuItem.IsActive,
	)

	if err != nil {
		return nil, err
	}

	return &menuItem, nil
}

// ListByRestaurant lists menu items for a restaurant
func (r *RepositoryImpl) ListByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]*MenuItem, error) {
	query := `
		SELECT id, restaurant_id, name, description, category, base_price, tax_category_id, is_active
		FROM menu_items
		WHERE restaurant_id = $1 AND is_active = true AND deleted_at IS NULL
		ORDER BY category, name
	`

	rows, err := r.db.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menuItems []*MenuItem
	for rows.Next() {
		var menuItem MenuItem
		err := rows.Scan(
			&menuItem.ID,
			&menuItem.RestaurantID,
			&menuItem.Name,
			&menuItem.Description,
			&menuItem.Category,
			&menuItem.BasePrice,
			&menuItem.TaxCategoryID,
			&menuItem.IsActive,
		)
		if err != nil {
			return nil, err
		}
		menuItems = append(menuItems, &menuItem)
	}

	return menuItems, nil
}

// Update updates a menu item
func (r *RepositoryImpl) Update(ctx context.Context, menuItem *MenuItem) error {
	query := `
		UPDATE menu_items
		SET name = $2, description = $3, category = $4, base_price = $5, tax_category_id = $6, updated_at = $7
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		menuItem.ID,
		menuItem.Name,
		menuItem.Description,
		menuItem.Category,
		menuItem.BasePrice,
		menuItem.TaxCategoryID,
		time.Now().UTC(),
	)

	return err
}

// Delete soft deletes a menu item
func (r *RepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE menu_items
		SET deleted_at = $1, is_active = false
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, time.Now().UTC(), id)
	return err
}
