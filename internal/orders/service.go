package orders

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service handles order operations
type Service struct {
	repo Repository
}

// New creates a new orders service
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// Order represents an order
type Order struct {
	ID               uuid.UUID
	CustomerID       uuid.UUID
	RestaurantID     uuid.UUID
	RegionID         uuid.UUID
	DeliveryAddressID uuid.UUID
	WorkflowState    string
	Subtotal         int
	DeliveryFee      int
	Tax              int
	Total            int
	CurrencyCode     string
	ScheduledFor     *string
	CreatedAt        string
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID         uuid.UUID
	OrderID    uuid.UUID
	MenuItemID uuid.UUID
	Quantity   int
	UnitPrice  int
}

// CreateOrderRequest represents a create order request
type CreateOrderRequest struct {
	RestaurantID      uuid.UUID
	DeliveryAddressID uuid.UUID
	Items            []CreateOrderItem
	ScheduledFor      *string
}

// CreateOrderItem represents an item to add to an order
type CreateOrderItem struct {
	MenuItemID uuid.UUID
	Quantity   int
}

// CreateOrder creates a new order
func (s *Service) CreateOrder(ctx context.Context, customerID uuid.UUID, req *CreateOrderRequest) (*Order, error) {
	// TODO: Implement order creation logic
	// - Validate restaurant is active and approved
	// - Validate delivery address belongs to customer
	// - Calculate subtotal from menu items
	// - Calculate delivery fee based on region config
	// - Calculate tax using tax service
	// - Create order with workflow state "pending"
	// - Create order items
	// - Emit order created event for background processing

	order := &Order{
		ID:               uuid.New(),
		CustomerID:       customerID,
		RestaurantID:     req.RestaurantID,
		RegionID:         uuid.New(), // TODO: Get from restaurant
		DeliveryAddressID: req.DeliveryAddressID,
		WorkflowState:    "pending",
		Subtotal:         0,
		DeliveryFee:      0,
		Tax:              0,
		Total:            0,
		CurrencyCode:     "USD",
		ScheduledFor:     req.ScheduledFor,
		CreatedAt:        "now",
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return order, nil
}

// GetOrder retrieves an order by ID
func (s *Service) GetOrder(ctx context.Context, id uuid.UUID) (*Order, error) {
	return s.repo.GetByID(ctx, id)
}

// ListCustomerOrders lists orders for a customer
func (s *Service) ListCustomerOrders(ctx context.Context, customerID uuid.UUID) ([]*Order, error) {
	return s.repo.ListByCustomer(ctx, customerID)
}

// ListRestaurantOrders lists orders for a restaurant
func (s *Service) ListRestaurantOrders(ctx context.Context, restaurantID uuid.UUID) ([]*Order, error) {
	return s.repo.ListByRestaurant(ctx, restaurantID)
}

// CancelOrder cancels an order
func (s *Service) CancelOrder(ctx context.Context, id uuid.UUID, customerID uuid.UUID) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order.CustomerID != customerID {
		return fmt.Errorf("order does not belong to customer")
	}

	// TODO: Validate order can be cancelled (workflow state check)
	return s.repo.UpdateWorkflowState(ctx, id, "cancelled")
}

// UpdateOrderState updates an order's workflow state (internal use)
func (s *Service) UpdateOrderState(ctx context.Context, id uuid.UUID, state string) error {
	return s.repo.UpdateWorkflowState(ctx, id, state)
}
