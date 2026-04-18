# Tax

## Overview

This document describes the tax model, inclusive vs exclusive tax modes, rounding rules, tax snapshot strategy, item vs delivery tax rules, and future support for jurisdiction-specific complexity.

## Tax Model Overview

The tax system is designed to be flexible, data-driven, and region-aware to support multiple tax jurisdictions worldwide.

### Core Concepts

1. **Tax Zones**: Geographic areas with specific tax rules (countries, states/provinces, cities)
2. **Tax Categories**: Product categories for tax classification (food, beverages, alcohol, delivery)
3. **Tax Rates**: Specific tax rates per tax zone and tax category
4. **Pricing Modes**: Tax-inclusive vs tax-exclusive pricing
5. **Tax Snapshots**: Tax calculations preserved at order time for historical accuracy

### Tax Data Model

```
Region
  └─ Tax Zone
      └─ Tax Rate (per Tax Category)
          └─ Tax Category
```

## Inclusive vs Exclusive Tax

### Tax-Exclusive Pricing

**Definition**: Prices displayed exclude tax, tax added at checkout

**Example**:
- Item price: $10.00
- Tax rate: 10%
- Tax amount: $1.00
- Total: $11.00

**Calculation**:
```
tax_amount = subtotal_amount * (rate_bps / 10000)
total_amount = subtotal_amount + tax_amount
```

**Regions**: United States (most states), Canada (most provinces)

### Tax-Inclusive Pricing

**Definition**: Prices displayed include tax, tax embedded in price

**Example**:
- Item price: $10.00 (includes tax)
- Tax rate: 10%
- Tax amount: $0.91
- Subtotal (pre-tax): $9.09

**Calculation**:
```
subtotal_amount = total_amount / (1 + rate_bps / 10000)
tax_amount = total_amount - subtotal_amount
```

**Regions**: United Kingdom (VAT), European Union (VAT), Australia (GST)

### Pricing Mode Configuration

Pricing mode configured per tax rate:

```sql
CREATE TABLE tax_rates (
    ...
    is_inclusive BOOLEAN NOT NULL DEFAULT false,
    ...
);
```

Per-region default can be configured in region config:
```sql
CREATE TABLE region_configs (
    ...
    pricing_mode VARCHAR(20) DEFAULT 'tax_exclusive',
    ...
);
```

## Rounding Rules

### Rounding Strategy

We use **banker's rounding** (round half to even) for tax calculations to minimize bias.

**Implementation**:
```go
import "math"

func roundTax(amount int) int {
    // amount is in minor units (cents)
    // Round to nearest cent using banker's rounding
    return int(math.RoundToEven(float64(amount)))
}
```

### Line Item Rounding

Tax calculated per line item, then rounded:
```
line_tax = round(line_subtotal * rate)
```

### Order Total Rounding

Total tax is sum of rounded line taxes:
```
total_tax = sum(line_tax_1, line_tax_2, ...)
```

### Delivery Fee Tax Rounding

Delivery fee tax calculated and rounded separately:
```
delivery_tax = round(delivery_fee * rate)
```

### Grand Total

Grand total is sum of all components:
```
grand_total = subtotal + delivery_fee + total_tax
```

## Tax Snapshot Strategy

### Why Snapshots?

Tax rates change over time. Historical orders must preserve the tax calculation at the time of order to ensure:
- Accurate financial reporting
- Invoice/receipt accuracy
- Compliance with tax regulations
- Audit trail integrity

### Snapshot Implementation

#### Order-Level Snapshot

```sql
CREATE TABLE orders (
    ...
    tax_amount INTEGER NOT NULL,
    pricing_mode VARCHAR(20) NOT NULL,
    tax_zone_id UUID,
    ...
);
```

#### Line-Level Snapshot

```sql
CREATE TABLE order_items (
    ...
    line_tax_amount INTEGER NOT NULL,
    tax_category_id UUID,
    ...
);
```

#### Tax Breakdown Snapshot

```sql
CREATE TABLE order_tax_lines (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL,
    tax_category_id UUID,
    tax_rate_name VARCHAR(100) NOT NULL,
    tax_rate_bps INTEGER NOT NULL,
    is_inclusive BOOLEAN NOT NULL,
    taxable_amount INTEGER NOT NULL,
    tax_amount INTEGER NOT NULL,
    currency_code VARCHAR(3) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);
```

### Snapshot Creation

When order is created, tax service:

1. Calculates tax based on current tax configuration
2. Creates order_tax_lines records with full breakdown
3. Stores tax_amount on order
4. Stores pricing_mode on order
5. Stores line_tax_amount on each order_item

### Snapshot Immutability

Once created, tax snapshots are never modified:
- Tax rate changes do not affect historical orders
- Tax zone changes do not affect historical orders
- Tax category changes do not affect historical orders
- Pricing mode changes do not affect historical orders

### Historical Query Example

```go
// Get order with tax breakdown as calculated at order time
order, _ := orderRepo.GetByID(ctx, orderID)
taxLines, _ := taxRepo.GetOrderTaxLines(ctx, orderID)

// Display invoice with original tax calculation
for _, line := range taxLines {
    fmt.Printf("%s: %d%% of %d = %d\n", 
        line.TaxRateName, 
        line.TaxRateBps / 100, 
        line.TaxableAmount, 
        line.TaxAmount)
}
```

## Item vs Delivery Tax Rules

### Item Taxation

**Rule**: Menu items are taxed based on their tax category

**Configuration**:
```sql
CREATE TABLE menu_items (
    ...
    tax_category_id UUID REFERENCES tax_categories(id),
    ...
);
```

**Calculation**:
```go
for _, item := range order.Items {
    taxRate := getTaxRate(item.TaxCategoryID, order.TaxZoneID)
    itemTax = calculateTax(item.LineSubtotal, taxRate)
}
```

### Delivery Fee Taxation

**Rule**: Delivery fee taxation is configurable per tax rate

**Configuration**:
```sql
CREATE TABLE tax_rates (
    ...
    applies_to_delivery_fee BOOLEAN NOT NULL DEFAULT false,
    ...
);
```

**Region Config Override**:
```sql
CREATE TABLE region_configs (
    ...
    delivery_fee_taxable_default BOOLEAN NOT NULL DEFAULT true,
    ...
);
```

**Calculation**:
```go
// Find tax rate that applies to delivery fee
deliveryTaxRate := getDeliveryTaxRate(order.TaxZoneID)
if deliveryTaxRate != nil && deliveryTaxRate.AppliesToDeliveryFee {
    deliveryTax = calculateTax(order.DeliveryFeeAmount, deliveryTaxRate)
}
```

### Tax Category for Delivery

Delivery fee typically has its own tax category:
```sql
INSERT INTO tax_categories (code, name) 
VALUES ('delivery', 'Delivery Fee');

INSERT INTO tax_rates (tax_zone_id, tax_category_id, name, rate_bps, applies_to_delivery_fee)
VALUES (tax_zone_id, delivery_category_id, 'Delivery Tax', 2000, true);
```

### Examples

**United States**:
- Food items: taxable in some states, exempt in others
- Delivery fee: often taxable
- Alcohol: always taxable

**United Kingdom**:
- Food items: zero-rated (0% VAT)
- Delivery fee: standard-rated (20% VAT)
- Alcohol: standard-rated (20% VAT)

**Canada**:
- Food items: exempt (0% GST in most provinces)
- Delivery fee: taxable (5% GST + provincial tax)
- Alcohol: taxable

## Tax Calculation Service

### Service Interface

```go
type TaxService interface {
    Calculate(ctx context.Context, taxZoneID uuid.UUID, items []TaxableItem, deliveryFee int) (*TaxCalculation, error)
    GetTaxZoneForAddress(ctx context.Context, address *Address) (uuid.UUID, error)
    ValidateTaxConfiguration(ctx context.Context, regionID uuid.UUID) error
}

type TaxableItem struct {
    Amount         int
    TaxCategoryID  uuid.UUID
}

type TaxCalculation struct {
    SubtotalAmount   int
    DeliveryFeeAmount int
    TaxAmount        int
    TotalAmount      int
    TaxLines         []TaxLine
    PricingMode      string
}

type TaxLine struct {
    TaxCategoryID   uuid.UUID
    TaxRateName     string
    TaxRateBPS      int
    IsInclusive     bool
    TaxableAmount   int
    TaxAmount       int
}
```

### Calculation Algorithm

```go
func (s *TaxService) Calculate(ctx context.Context, taxZoneID uuid.UUID, items []TaxableItem, deliveryFee int) (*TaxCalculation, error) {
    // Fetch all applicable tax rates for tax zone
    taxRates, err := s.taxRateRepo.GetByTaxZone(ctx, taxZoneID)
    if err != nil {
        return nil, err
    }
    
    var taxLines []TaxLine
    var totalTax int
    var subtotal int
    
    // Calculate item taxes
    for _, item := range items {
        subtotal += item.Amount
        
        taxRate := findTaxRate(taxRates, item.TaxCategoryID)
        if taxRate == nil {
            continue // No tax for this category
        }
        
        taxAmount := s.calculateTax(item.Amount, taxRate)
        totalTax += taxAmount
        
        taxLines = append(taxLines, TaxLine{
            TaxCategoryID: item.TaxCategoryID,
            TaxRateName:   taxRate.Name,
            TaxRateBPS:    taxRate.RateBPS,
            IsInclusive:  taxRate.IsInclusive,
            TaxableAmount: item.Amount,
            TaxAmount:    taxAmount,
        })
    }
    
    // Calculate delivery fee tax
    deliveryTaxRate := findDeliveryTaxRate(taxRates)
    if deliveryTaxRate != nil && deliveryFee > 0 {
        deliveryTax := s.calculateTax(deliveryFee, deliveryTaxRate)
        totalTax += deliveryTax
        
        taxLines = append(taxLines, TaxLine{
            TaxCategoryID: deliveryTaxRate.TaxCategoryID,
            TaxRateName:   deliveryTaxRate.Name,
            TaxRateBPS:    deliveryTaxRate.RateBPS,
            IsInclusive:  deliveryTaxRate.IsInclusive,
            TaxableAmount: deliveryFee,
            TaxAmount:    deliveryTax,
        })
    }
    
    // Determine pricing mode (default to exclusive if mixed)
    pricingMode := s.determinePricingMode(taxRates)
    
    // Adjust for inclusive pricing
    if pricingMode == "tax_inclusive" {
        subtotal = s.calculateSubtotalFromInclusive(subtotal + deliveryFee, totalTax)
    }
    
    return &TaxCalculation{
        SubtotalAmount:   subtotal,
        DeliveryFeeAmount: deliveryFee,
        TaxAmount:        totalTax,
        TotalAmount:      subtotal + deliveryFee + totalTax,
        TaxLines:         taxLines,
        PricingMode:      pricingMode,
    }, nil
}

func (s *TaxService) calculateTax(amount int, taxRate *TaxRate) int {
    if taxRate.IsInclusive {
        // Tax is embedded in amount
        taxAmount := amount - int(float64(amount)/(1+float64(taxRate.RateBPS)/10000))
        return roundTax(taxAmount)
    } else {
        // Tax added to amount
        taxAmount := int(float64(amount) * float64(taxRate.RateBPS) / 10000)
        return roundTax(taxAmount)
    }
}
```

### Tax Zone Resolution

```go
func (s *TaxService) GetTaxZoneForAddress(ctx context.Context, address *Address) (uuid.UUID, error) {
    // Try to match by postal code pattern first
    taxZone, err := s.taxZoneRepo.GetByPostalCode(ctx, address.PostalCode, address.CountryCode)
    if err == nil {
        return taxZone.ID, nil
    }
    
    // Try to match by city
    taxZone, err = s.taxZoneRepo.GetByCity(ctx, address.City, address.StateOrProvince, address.CountryCode)
    if err == nil {
        return taxZone.ID, nil
    }
    
    // Try to match by state/province
    taxZone, err = s.taxZoneRepo.GetByState(ctx, address.StateOrProvince, address.CountryCode)
    if err == nil {
        return taxZone.ID, nil
    }
    
    // Fallback to country-wide tax zone
    taxZone, err = s.taxZoneRepo.GetByCountry(ctx, address.CountryCode)
    if err == nil {
        return taxZone.ID, nil
    }
    
    return uuid.Nil, ErrTaxZoneNotFound
}
```

## Future Support for Jurisdiction-Specific Complexity

### Compound Tax Rates

**Current**: Single tax rate per category per zone

**Future**: Support for compound tax rates (e.g., Canada: GST + provincial tax)

**Implementation**:
```sql
-- Add priority field for ordering
ALTER TABLE tax_rates ADD COLUMN priority INTEGER NOT NULL DEFAULT 0;

-- Multiple rates can apply to same category
-- Calculate in priority order
```

### Tax Exemptions

**Current**: All items in category taxed at same rate

**Future**: Support for tax-exempt items or customers

**Implementation**:
```sql
CREATE TABLE tax_exemptions (
    id UUID PRIMARY KEY,
    entity_type VARCHAR(50), -- 'menu_item', 'customer'
    entity_id UUID,
    tax_zone_id UUID,
    reason VARCHAR(100),
    effective_from TIMESTAMP WITH TIME ZONE,
    effective_to TIMESTAMP WITH TIME ZONE
);
```

### Tax Thresholds

**Current**: Flat tax rate on all amounts

**Future**: Support for tax thresholds (e.g., tax only applies above certain amount)

**Implementation**:
```sql
ALTER TABLE tax_rates ADD COLUMN threshold_amount INTEGER;
ALTER TABLE tax_rates ADD COLUMN threshold_rate_bps INTEGER;

-- Apply different rate above threshold
if amount > threshold_amount {
    tax = threshold_amount * rate + (amount - threshold_amount) * threshold_rate
}
```

### Tax Holidays

**Current**: No support for temporary tax changes

**Future**: Support for tax holidays with effective date ranges

**Implementation**:
```sql
-- Already supported via effective_from/effective_to
INSERT INTO tax_rates (...)
VALUES (..., effective_from = '2024-12-01', effective_to = '2024-12-31');
```

### Reverse Charge

**Current**: Not supported

**Future**: Support for reverse charge mechanism (B2B cross-border)

**Implementation**:
```sql
ALTER TABLE tax_rates ADD COLUMN is_reverse_charge BOOLEAN DEFAULT false;
```

### Digital Services Tax

**Current**: Not applicable (food delivery)

**Future**: Support for digital services tax if expanded to other services

### Tax Reporting

**Current**: Basic tax breakdown stored

**Future**: Enhanced tax reporting by jurisdiction, date range, tax type

**Implementation**:
```sql
-- Reporting views
CREATE VIEW tax_report_by_zone AS
SELECT 
    tz.name as tax_zone,
    DATE_TRUNC('month', o.created_at) as month,
    SUM(o.tax_amount) as total_tax
FROM orders o
JOIN tax_zones tz ON o.tax_zone_id = tz.id
GROUP BY tz.name, DATE_TRUNC('month', o.created_at);
```

## Tax Configuration Examples

### United States - California

```sql
-- Region
INSERT INTO regions (code, name, country_code, timezone, currency_code)
VALUES ('US-CA', 'California', 'US', 'America/Los_Angeles', 'USD');

-- Tax Zone
INSERT INTO tax_zones (code, name, country_code, state_or_province)
VALUES ('US-CA', 'California', 'US', 'CA');

-- Tax Categories
INSERT INTO tax_categories (code, name) VALUES
('food', 'Food and Beverages'),
('alcohol', 'Alcoholic Beverages'),
('delivery', 'Delivery Fee');

-- Tax Rates
INSERT INTO tax_rates (tax_zone_id, tax_category_id, name, rate_bps, is_inclusive, applies_to_delivery_fee)
VALUES
(tax_zone_id, food_category_id, 'California Sales Tax', 825, false, true),
(tax_zone_id, alcohol_category_id, 'California Alcohol Tax', 825, false, false),
(tax_zone_id, delivery_category_id, 'California Delivery Tax', 825, false, true);
```

### United Kingdom - VAT

```sql
-- Region
INSERT INTO regions (code, name, country_code, timezone, currency_code)
VALUES ('GB', 'United Kingdom', 'GB', 'Europe/London', 'GBP');

-- Tax Zone
INSERT INTO tax_zones (code, name, country_code)
VALUES ('GB', 'United Kingdom', 'GB');

-- Tax Categories
INSERT INTO tax_categories (code, name) VALUES
('food', 'Food and Beverages'),
('alcohol', 'Alcoholic Beverages'),
('delivery', 'Delivery Fee');

-- Tax Rates
INSERT INTO tax_rates (tax_zone_id, tax_category_id, name, rate_bps, is_inclusive, applies_to_delivery_fee)
VALUES
(tax_zone_id, food_category_id, 'Zero Rated Food', 0, true, false),
(tax_zone_id, alcohol_category_id, 'Standard Rate', 2000, true, false),
(tax_zone_id, delivery_category_id, 'Standard Rate', 2000, true, true);
```

### Canada - GST + Provincial

```sql
-- Region
INSERT INTO regions (code, name, country_code, timezone, currency_code)
VALUES ('CA-ON', 'Ontario', 'CA', 'America/Toronto', 'CAD');

-- Tax Zone
INSERT INTO tax_zones (code, name, country_code, state_or_province)
VALUES ('CA-ON', 'Ontario', 'CA', 'ON');

-- Tax Categories
INSERT INTO tax_categories (code, name) VALUES
('food', 'Food and Beverages'),
('alcohol', 'Alcoholic Beverages'),
('delivery', 'Delivery Fee');

-- Tax Rates (GST first, then HST)
INSERT INTO tax_rates (tax_zone_id, tax_category_id, name, rate_bps, is_inclusive, applies_to_delivery_fee, priority)
VALUES
(tax_zone_id, food_category_id, 'GST', 500, true, false, 1),
(tax_zone_id, food_category_id, 'Ontario HST', 800, true, false, 2),
(tax_zone_id, delivery_category_id, 'GST', 500, true, true, 1),
(tax_zone_id, delivery_category_id, 'Ontario HST', 800, true, true, 2);
```

## Tax Validation

### Configuration Validation

```go
func (s *TaxService) ValidateTaxConfiguration(ctx context.Context, regionID uuid.UUID) error {
    // Check that region has at least one tax zone
    taxZones, err := s.taxZoneRepo.GetByRegion(ctx, regionID)
    if err != nil || len(taxZones) == 0 {
        return ErrNoTaxZonesForRegion
    }
    
    // Check that each tax zone has tax rates
    for _, zone := range taxZones {
        rates, err := s.taxRateRepo.GetByTaxZone(ctx, zone.ID)
        if err != nil || len(rates) == 0 {
            return fmt.Errorf("tax zone %s has no tax rates", zone.Code)
        }
        
        // Check for overlapping effective dates
        if hasOverlappingDates(rates) {
            return ErrOverlappingTaxRates
        }
    }
    
    return nil
}
```

### Runtime Validation

```go
func (s *TaxService) Calculate(ctx context.Context, ...) (*TaxCalculation, error) {
    // Validate tax zone exists
    taxZone, err := s.taxZoneRepo.GetByID(ctx, taxZoneID)
    if err != nil {
        return nil, ErrTaxZoneNotFound
    }
    
    // Validate tax zone is active
    if !taxZone.IsActive {
        return nil, ErrTaxZoneInactive
    }
    
    // Validate tax rates are effective
    for _, rate := range taxRates {
        now := time.Now()
        if now.Before(rate.EffectiveFrom) || (rate.EffectiveTo != nil && now.After(*rate.EffectiveTo)) {
            continue // Skip inactive rates
        }
    }
    
    // ... calculation
}
```

## Tax Administration

### Admin API for Tax Configuration

**Create Tax Rate**:
```http
POST /v1/admin/tax-rates
{
  "tax_zone_id": "uuid",
  "tax_category_id": "uuid",
  "name": "California Sales Tax",
  "rate_bps": 825,
  "is_inclusive": false,
  "applies_to_delivery_fee": true,
  "effective_from": "2024-01-01T00:00:00Z"
}
```

**Update Tax Rate**:
```http
PATCH /v1/admin/tax-rates/{id}
{
  "rate_bps": 850,
  "effective_to": "2024-12-31T23:59:59Z"
}
```

**Deactivate Tax Rate**:
```http
DELETE /v1/admin/tax-rates/{id}
```

### Audit Logging

All tax configuration changes create audit logs:
```sql
INSERT INTO audit_logs (actor_id, actor_role, action, entity_type, entity_id, metadata)
VALUES (admin_id, 'admin', 'tax_rate_updated', 'tax_rate', tax_rate_id, '{"previous_rate_bps": 825, "new_rate_bps": 850}');
```

## Summary

The tax system provides:
- Flexible, data-driven tax configuration
- Support for inclusive and exclusive tax modes
- Consistent rounding rules
- Tax snapshot immutability for historical accuracy
- Item and delivery fee taxation
- Region-aware tax zone resolution
- Server-side tax calculation only
- Future support for complex jurisdiction rules
- Admin API for tax management
- Audit logging for tax changes
