package delivery

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service handles delivery operations
type Service struct {
	repo Repository
}

// New creates a new delivery service
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// Delivery represents a delivery assignment
type Delivery struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	CourierID   *uuid.UUID
	Status      string
	PickedUpAt  *string
	DeliveredAt *string
}

// Location represents a geographic location
type Location struct {
	Latitude  float64
	Longitude float64
}

// DeliveryFeeRequest represents a delivery fee calculation request
type DeliveryFeeRequest struct {
	PickupLocation   Location
	DeliveryLocation Location
	BaseFee          int
	DistanceFeePerKm int
	SurgeMultiplier  float64
	IsSurgeActive    bool
}

// DeliveryFeeResponse represents a delivery fee calculation response
type DeliveryFeeRequestResponse struct {
	BaseFee         int
	DistanceKm      float64
	DistanceFee     int
	Surcharge       int
	TotalFee        int
	SurgeMultiplier float64
	IsSurgeActive   bool
}

// CreateDelivery creates a new delivery assignment
func (s *Service) CreateDelivery(ctx context.Context, orderID uuid.UUID) (*Delivery, error) {
	delivery := &Delivery{
		ID:      uuid.New(),
		OrderID: orderID,
		Status:  "pending",
	}

	if err := s.repo.Create(ctx, delivery); err != nil {
		return nil, fmt.Errorf("failed to create delivery: %w", err)
	}

	return delivery, nil
}

// GetDelivery retrieves a delivery by ID
func (s *Service) GetDelivery(ctx context.Context, id uuid.UUID) (*Delivery, error) {
	return s.repo.GetByID(ctx, id)
}

// ListCourierDeliveries lists deliveries for a courier
func (s *Service) ListCourierDeliveries(ctx context.Context, courierID uuid.UUID) ([]*Delivery, error) {
	return s.repo.ListByCourier(ctx, courierID)
}

// AssignCourier assigns a courier to a delivery
func (s *Service) AssignCourier(ctx context.Context, deliveryID uuid.UUID, courierID uuid.UUID) error {
	return s.repo.AssignCourier(ctx, deliveryID, courierID)
}

// UpdateStatus updates a delivery's status
func (s *Service) UpdateStatus(ctx context.Context, deliveryID uuid.UUID, status string) error {
	return s.repo.UpdateStatus(ctx, deliveryID, status)
}

// MarkPickedUp marks a delivery as picked up
func (s *Service) MarkPickedUp(ctx context.Context, deliveryID uuid.UUID) error {
	return s.repo.MarkPickedUp(ctx, deliveryID)
}

// MarkDelivered marks a delivery as delivered
func (s *Service) MarkDelivered(ctx context.Context, deliveryID uuid.UUID) error {
	return s.repo.MarkDelivered(ctx, deliveryID)
}
