package payments

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service handles payment operations
type Service struct {
	repo Repository
}

// New creates a new payments service
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// Payment represents a payment
type Payment struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	Amount      int
	CurrencyCode string
	Status      string
	PaymentMethod string
	PaymentProviderID *string
}

// CreatePaymentRequest represents a create payment request
type CreatePaymentRequest struct {
	OrderID       uuid.UUID
	Amount        int
	PaymentMethod string
}

// CreatePayment creates a new payment
func (s *Service) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*Payment, error) {
	payment := &Payment{
		ID:            uuid.New(),
		OrderID:       req.OrderID,
		Amount:        req.Amount,
		CurrencyCode:  "USD", // TODO: Get from order
		Status:        "pending",
		PaymentMethod: req.PaymentMethod,
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return payment, nil
}

// GetPayment retrieves a payment by ID
func (s *Service) GetPayment(ctx context.Context, id uuid.UUID) (*Payment, error) {
	return s.repo.GetByID(ctx, id)
}

// ListOrderPayments lists payments for an order
func (s *Service) ListOrderPayments(ctx context.Context, orderID uuid.UUID) ([]*Payment, error) {
	return s.repo.ListByOrder(ctx, orderID)
}

// UpdatePaymentStatus updates a payment's status
func (s *Service) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status string, providerID *string) error {
	return s.repo.UpdateStatus(ctx, id, status, providerID)
}
