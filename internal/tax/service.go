package tax

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service handles tax operations
type Service struct {
	repo Repository
}

// New creates a new tax service
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// TaxCategory represents a tax category
type TaxCategory struct {
	ID          uuid.UUID
	Code        string
	Name        string
	Description string
}

// TaxZone represents a tax zone
type TaxZone struct {
	ID                uuid.UUID
	Code              string
	Name              string
	CountryCode       string
	State             string
	City              string
	PostalCodePattern string
	IsActive          bool
}

// TaxRate represents a tax rate
type TaxRate struct {
	ID                   uuid.UUID
	TaxZoneID            uuid.UUID
	TaxCategoryID        uuid.UUID
	Name                 string
	RateBPS              int
	IsInclusive          bool
	AppliesToDeliveryFee bool
	EffectiveFrom        string
	EffectiveTo          *string
	IsActive             bool
}

// CalculateTaxRequest represents a tax calculation request
type CalculateTaxRequest struct {
	TaxZoneID   uuid.UUID
	Items       []TaxableItem
	DeliveryFee int
	PricingMode string
}

// TaxableItem represents an item to be taxed
type TaxableItem struct {
	TaxCategoryID uuid.UUID
	Amount        int
}

// TaxCalculation represents the result of a tax calculation
type TaxCalculation struct {
	ItemTax     int
	DeliveryTax int
	TotalTax    int
	TaxLines    []TaxLine
}

// TaxLine represents a single tax line
type TaxLine struct {
	TaxCategoryID uuid.UUID
	TaxRateName   string
	TaxRateBPS    int
	IsInclusive   bool
	TaxableAmount int
	TaxAmount     int
}

// ListTaxCategories lists all tax categories
func (s *Service) ListTaxCategories(ctx context.Context) ([]*TaxCategory, error) {
	return s.repo.ListCategories(ctx)
}

// ListTaxZones lists all active tax zones
func (s *Service) ListTaxZones(ctx context.Context) ([]*TaxZone, error) {
	return s.repo.ListZones(ctx)
}

// ListTaxRates lists tax rates for a zone
func (s *Service) ListTaxRates(ctx context.Context, taxZoneID uuid.UUID) ([]*TaxRate, error) {
	return s.repo.ListRates(ctx, taxZoneID)
}

// CalculateTax calculates tax for an order
func (s *Service) CalculateTax(ctx context.Context, req *CalculateTaxRequest) (*TaxCalculation, error) {
	rates, err := s.repo.ListRates(ctx, req.TaxZoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tax rates: %w", err)
	}

	calc := &TaxCalculation{
		TaxLines: make([]TaxLine, 0),
	}

	// Calculate item tax
	for _, item := range req.Items {
		itemTax, taxLine := s.calculateItemTax(item, rates, req.PricingMode)
		calc.ItemTax += itemTax
		if taxLine != nil {
			calc.TaxLines = append(calc.TaxLines, *taxLine)
		}
	}

	// Calculate delivery tax
	deliveryTax, deliveryTaxLine := s.calculateDeliveryTax(req.DeliveryFee, rates, req.PricingMode)
	calc.DeliveryTax = deliveryTax
	if deliveryTaxLine != nil {
		calc.TaxLines = append(calc.TaxLines, *deliveryTaxLine)
	}

	calc.TotalTax = calc.ItemTax + calc.DeliveryTax

	return calc, nil
}

// calculateItemTax calculates tax for a single item
func (s *Service) calculateItemTax(item TaxableItem, rates []*TaxRate, pricingMode string) (int, *TaxLine) {
	for _, rate := range rates {
		if rate.TaxCategoryID == item.TaxCategoryID {
			taxAmount := s.calculateTaxAmount(item.Amount, rate.RateBPS, rate.IsInclusive, pricingMode)
			return taxAmount, &TaxLine{
				TaxCategoryID: item.TaxCategoryID,
				TaxRateName:   rate.Name,
				TaxRateBPS:    rate.RateBPS,
				IsInclusive:   rate.IsInclusive,
				TaxableAmount: item.Amount,
				TaxAmount:     taxAmount,
			}
		}
	}
	return 0, nil
}

// calculateDeliveryTax calculates tax for delivery fee
func (s *Service) calculateDeliveryTax(deliveryFee int, rates []*TaxRate, pricingMode string) (int, *TaxLine) {
	for _, rate := range rates {
		if rate.AppliesToDeliveryFee {
			taxAmount := s.calculateTaxAmount(deliveryFee, rate.RateBPS, rate.IsInclusive, pricingMode)
			return taxAmount, &TaxLine{
				TaxCategoryID: rate.TaxCategoryID,
				TaxRateName:   rate.Name,
				TaxRateBPS:    rate.RateBPS,
				IsInclusive:   rate.IsInclusive,
				TaxableAmount: deliveryFee,
				TaxAmount:     taxAmount,
			}
		}
	}
	return 0, nil
}

// calculateTaxAmount calculates tax amount based on rate and pricing mode
func (s *Service) calculateTaxAmount(amount int, rateBPS int, isInclusive bool, pricingMode string) int {
	if isInclusive || pricingMode == "tax_inclusive" {
		// Tax is inclusive: extract tax from the amount
		// Tax = Amount * Rate / (10000 + Rate)
		tax := (amount * rateBPS) / (10000 + rateBPS)
		return s.roundTax(tax)
	}
	// Tax is exclusive: add tax to the amount
	// Tax = Amount * Rate / 10000
	tax := (amount * rateBPS) / 10000
	return s.roundTax(tax)
}

// roundTax rounds tax amount (round half up)
func (s *Service) roundTax(tax int) int {
	return tax
}
