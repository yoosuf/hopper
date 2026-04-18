package tax

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for tax data access
type Repository interface {
	ListCategories(ctx context.Context) ([]*TaxCategory, error)
	ListZones(ctx context.Context) ([]*TaxZone, error)
	ListRates(ctx context.Context, taxZoneID uuid.UUID) ([]*TaxRate, error)
}

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRepository creates a new tax repository
func NewRepository(db *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

// ListCategories lists all tax categories
func (r *RepositoryImpl) ListCategories(ctx context.Context) ([]*TaxCategory, error) {
	query := `
		SELECT id, code, name, description
		FROM tax_categories
		ORDER BY code
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*TaxCategory
	for rows.Next() {
		var cat TaxCategory
		err := rows.Scan(
			&cat.ID,
			&cat.Code,
			&cat.Name,
			&cat.Description,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, &cat)
	}

	return categories, nil
}

// ListZones lists all active tax zones
func (r *RepositoryImpl) ListZones(ctx context.Context) ([]*TaxZone, error) {
	query := `
		SELECT id, code, name, country_code, state_or_province, city, postal_code_pattern, is_active
		FROM tax_zones
		WHERE is_active = true AND deleted_at IS NULL
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []*TaxZone
	for rows.Next() {
		var zone TaxZone
		err := rows.Scan(
			&zone.ID,
			&zone.Code,
			&zone.Name,
			&zone.CountryCode,
			&zone.State,
			&zone.City,
			&zone.PostalCodePattern,
			&zone.IsActive,
		)
		if err != nil {
			return nil, err
		}
		zones = append(zones, &zone)
	}

	return zones, nil
}

// ListRates lists all active tax rates for a zone
func (r *RepositoryImpl) ListRates(ctx context.Context, taxZoneID uuid.UUID) ([]*TaxRate, error) {
	query := `
		SELECT id, tax_zone_id, tax_category_id, name, rate_bps, is_inclusive, applies_to_delivery_fee, effective_from, effective_to, is_active
		FROM tax_rates
		WHERE tax_zone_id = $1 AND is_active = true AND deleted_at IS NULL
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query, taxZoneID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []*TaxRate
	for rows.Next() {
		var rate TaxRate
		err := rows.Scan(
			&rate.ID,
			&rate.TaxZoneID,
			&rate.TaxCategoryID,
			&rate.Name,
			&rate.RateBPS,
			&rate.IsInclusive,
			&rate.AppliesToDeliveryFee,
			&rate.EffectiveFrom,
			&rate.EffectiveTo,
			&rate.IsActive,
		)
		if err != nil {
			return nil, err
		}
		rates = append(rates, &rate)
	}

	return rates, nil
}
