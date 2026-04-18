package regions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service handles region operations
type Service struct {
	repo Repository
}

// New creates a new regions service
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// Region represents a region
type Region struct {
	ID           uuid.UUID
	Code         string
	Name         string
	CountryCode  string
	Timezone     string
	CurrencyCode string
	IsActive     bool
}

// RegionConfig represents region-specific configuration
type RegionConfig struct {
	RegionID                     uuid.UUID
	PlatformFeeBasisPoints      int
	DefaultDeliveryWindowMinutes int
	OrderActivationLeadMinutes  int
	AllowScheduledOrders        bool
	DeliveryFeeTaxableDefault   bool
}

// ListRegions lists all active regions
func (s *Service) ListRegions(ctx context.Context) ([]*Region, error) {
	return s.repo.ListActive(ctx)
}

// GetRegion retrieves a region by ID
func (s *Service) GetRegion(ctx context.Context, id uuid.UUID) (*Region, error) {
	region, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get region: %w", err)
	}
	return region, nil
}

// GetRegionConfig retrieves region configuration
func (s *Service) GetRegionConfig(ctx context.Context, regionID uuid.UUID) (*RegionConfig, error) {
	config, err := s.repo.GetConfig(ctx, regionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get region config: %w", err)
	}
	return config, nil
}
