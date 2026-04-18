package regions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for region data access
type Repository interface {
	ListActive(ctx context.Context) ([]*Region, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Region, error)
	GetConfig(ctx context.Context, regionID uuid.UUID) (*RegionConfig, error)
}

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRepository creates a new regions repository
func NewRepository(db *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

// ListActive lists all active regions
func (r *RepositoryImpl) ListActive(ctx context.Context) ([]*Region, error) {
	query := `
		SELECT id, code, name, country_code, timezone, currency_code, is_active
		FROM regions
		WHERE is_active = true AND deleted_at IS NULL
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regions []*Region
	for rows.Next() {
		var region Region
		err := rows.Scan(
			&region.ID,
			&region.Code,
			&region.Name,
			&region.CountryCode,
			&region.Timezone,
			&region.CurrencyCode,
			&region.IsActive,
		)
		if err != nil {
			return nil, err
		}
		regions = append(regions, &region)
	}

	return regions, nil
}

// GetByID retrieves a region by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*Region, error) {
	query := `
		SELECT id, code, name, country_code, timezone, currency_code, is_active
		FROM regions
		WHERE id = $1 AND deleted_at IS NULL
	`

	var region Region
	err := r.db.QueryRow(ctx, query, id).Scan(
		&region.ID,
		&region.Code,
		&region.Name,
		&region.CountryCode,
		&region.Timezone,
		&region.CurrencyCode,
		&region.IsActive,
	)

	if err != nil {
		return nil, err
	}

	return &region, nil
}

// GetConfig retrieves region configuration
func (r *RepositoryImpl) GetConfig(ctx context.Context, regionID uuid.UUID) (*RegionConfig, error) {
	query := `
		SELECT region_id, platform_fee_basis_points, default_delivery_window_minutes, order_activation_lead_minutes, allow_scheduled_orders, delivery_fee_taxable_default
		FROM region_configs
		WHERE region_id = $1
	`

	var config RegionConfig
	err := r.db.QueryRow(ctx, query, regionID).Scan(
		&config.RegionID,
		&config.PlatformFeeBasisPoints,
		&config.DefaultDeliveryWindowMinutes,
		&config.OrderActivationLeadMinutes,
		&config.AllowScheduledOrders,
		&config.DeliveryFeeTaxableDefault,
	)

	if err != nil {
		return nil, err
	}

	return &config, nil
}
