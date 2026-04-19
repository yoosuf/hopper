package orders

import (
	"context"
	"fmt"
	"time"

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
	ID                uuid.UUID
	CustomerID        uuid.UUID
	RestaurantID      uuid.UUID
	RegionID          uuid.UUID
	DeliveryAddressID uuid.UUID
	WorkflowState     string
	Subtotal          int
	DeliveryFee       int
	Tax               int
	Total             int
	CurrencyCode      string
	ScheduledFor      *string
	CreatedAt         string
}

// OrderWorkflowState represents the valid workflow states for an order
type OrderWorkflowState string

const (
	WorkflowStatePending   OrderWorkflowState = "pending"
	WorkflowStateConfirmed OrderWorkflowState = "confirmed"
	WorkflowStatePreparing OrderWorkflowState = "preparing"
	WorkflowStateReady     OrderWorkflowState = "ready"
	WorkflowStatePickedUp  OrderWorkflowState = "picked_up"
	WorkflowStateDelivered OrderWorkflowState = "delivered"
	WorkflowStateCancelled OrderWorkflowState = "cancelled"
)

// ValidStateTransitions defines valid state transitions
var ValidStateTransitions = map[OrderWorkflowState][]OrderWorkflowState{
	WorkflowStatePending:   {WorkflowStateConfirmed, WorkflowStateCancelled},
	WorkflowStateConfirmed: {WorkflowStatePreparing, WorkflowStateCancelled},
	WorkflowStatePreparing: {WorkflowStateReady},
	WorkflowStateReady:     {WorkflowStatePickedUp},
	WorkflowStatePickedUp:  {WorkflowStateDelivered},
	WorkflowStateDelivered: {},
	WorkflowStateCancelled: {},
}

// CanTransitionTo checks if a state transition is valid
func (s OrderWorkflowState) CanTransitionTo(newState OrderWorkflowState) bool {
	validTransitions, exists := ValidStateTransitions[s]
	if !exists {
		return false
	}
	for _, validState := range validTransitions {
		if validState == newState {
			return true
		}
	}
	return false
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
	Items             []CreateOrderItem
	ScheduledFor      *string
}

// CreateOrderItem represents an item to add to an order
type CreateOrderItem struct {
	MenuItemID uuid.UUID
	Quantity   int
}

// CreateOrder creates a new order
func (s *Service) CreateOrder(ctx context.Context, customerID uuid.UUID, req *CreateOrderRequest) (*Order, error) {
	// Validate restaurant is active and approved
	isActive, isApproved, currencyCode, err := s.repo.GetRestaurantStatus(ctx, req.RestaurantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get restaurant status: %w", err)
	}
	if !isActive {
		return nil, fmt.Errorf("restaurant is not active")
	}
	if !isApproved {
		return nil, fmt.Errorf("restaurant is not approved")
	}

	// Validate delivery address belongs to customer
	addressBelongsToCustomer, err := s.repo.ValidateAddressOwnership(ctx, req.DeliveryAddressID, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate address ownership: %w", err)
	}
	if !addressBelongsToCustomer {
		return nil, fmt.Errorf("delivery address does not belong to customer")
	}

	// Get region from restaurant for region scope enforcement
	regionID, err := s.repo.GetRestaurantRegion(ctx, req.RestaurantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get restaurant region: %w", err)
	}

	// Calculate subtotal from menu items
	subtotal := 0
	for _, item := range req.Items {
		price, err := s.repo.GetMenuItemPrice(ctx, item.MenuItemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get menu item price: %w", err)
		}
		subtotal += price * item.Quantity
	}

	// Calculate delivery fee (fixed for now, could be distance-based)
	deliveryFee := 500 // $5.00 in cents

	// Calculate tax (simplified: 10% of subtotal + delivery fee)
	tax := (subtotal + deliveryFee) * 10 / 100

	// Calculate total
	total := subtotal + deliveryFee + tax

	// Create order with workflow state "pending"
	order := &Order{
		ID:                uuid.New(),
		CustomerID:        customerID,
		RestaurantID:      req.RestaurantID,
		RegionID:          regionID,
		DeliveryAddressID: req.DeliveryAddressID,
		WorkflowState:     string(WorkflowStatePending),
		Subtotal:          subtotal,
		DeliveryFee:       deliveryFee,
		Tax:               tax,
		Total:             total,
		CurrencyCode:      currencyCode,
		ScheduledFor:      req.ScheduledFor,
		CreatedAt:         time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// TODO: Create order items
	// TODO: Emit order created event for background processing

	return order, nil
}

// GetOrder retrieves an order by ID
func (s *Service) GetOrder(ctx context.Context, id uuid.UUID) (*Order, error) {
	return s.repo.GetByID(ctx, id)
}

// ListCustomerOrders lists orders for a customer with pagination
func (s *Service) ListCustomerOrders(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]*Order, error) {
	return s.repo.ListByCustomer(ctx, customerID, limit, offset)
}

// ListRestaurantOrders lists orders for a restaurant with pagination
func (s *Service) ListRestaurantOrders(ctx context.Context, restaurantID uuid.UUID, limit, offset int) ([]*Order, error) {
	return s.repo.ListByRestaurant(ctx, restaurantID, limit, offset)
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

	// Validate order can be cancelled using workflow state machine
	currentState := OrderWorkflowState(order.WorkflowState)
	if !currentState.CanTransitionTo(WorkflowStateCancelled) {
		return fmt.Errorf("order cannot be cancelled from current state: %s", currentState)
	}

	return s.repo.UpdateWorkflowState(ctx, id, string(WorkflowStateCancelled))
}

// GetOrderHistory retrieves order history for a customer with pagination
func (s *Service) GetOrderHistory(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]*Order, error) {
	return s.repo.ListByCustomer(ctx, customerID, limit, offset)
}

// GetOrderHistoryStats retrieves order statistics for a customer
func (s *Service) GetOrderHistoryStats(ctx context.Context, customerID uuid.UUID) (map[string]interface{}, error) {
	return s.repo.GetCustomerStats(ctx, customerID)
}

// ReorderRequest represents a reorder request
type ReorderRequest struct {
	OriginalOrderID uuid.UUID
	NewAddressID    uuid.UUID
}

// Reorder creates a new order based on a previous order
func (s *Service) Reorder(ctx context.Context, customerID uuid.UUID, req *ReorderRequest) (*Order, error) {
	// Get original order
	originalOrder, err := s.repo.GetByID(ctx, req.OriginalOrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original order: %w", err)
	}

	if originalOrder.CustomerID != customerID {
		return nil, fmt.Errorf("order does not belong to customer")
	}

	// Get order items from original order
	orderItems, err := s.repo.GetOrderItems(ctx, req.OriginalOrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	// Create new order request with same items
	createReq := &CreateOrderRequest{
		RestaurantID:      originalOrder.RestaurantID,
		DeliveryAddressID: req.NewAddressID,
		Items:             make([]CreateOrderItem, len(orderItems)),
		ScheduledFor:      nil,
	}

	for i, item := range orderItems {
		createReq.Items[i] = CreateOrderItem{
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
		}
	}

	// Create the new order
	return s.CreateOrder(ctx, customerID, createReq)
}

// UpdateOrderState updates an order's workflow state (internal use)
func (s *Service) UpdateOrderState(ctx context.Context, id uuid.UUID, state string) error {
	return s.repo.UpdateWorkflowState(ctx, id, state)
}
