package orders

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for order data access
type Repository interface {
	Create(ctx context.Context, order *Order) error
	CreateWithItems(ctx context.Context, order *Order, items []*OrderItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	ListByCustomer(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]*Order, error)
	ListByCustomerWithItems(ctx context.Context, customerID uuid.UUID, limit, offset int) (map[uuid.UUID][]*OrderItem, []*Order, error)
	ListByRestaurant(ctx context.Context, restaurantID uuid.UUID, limit, offset int) ([]*Order, error)
	UpdateWorkflowState(ctx context.Context, id uuid.UUID, state string) error
	GetRestaurantRegion(ctx context.Context, restaurantID uuid.UUID) (uuid.UUID, error)
	GetRestaurantStatus(ctx context.Context, restaurantID uuid.UUID) (isActive bool, isApproved bool, currencyCode string, err error)
	ValidateAddressOwnership(ctx context.Context, addressID uuid.UUID, customerID uuid.UUID) (bool, error)
	GetMenuItemPrice(ctx context.Context, menuItemID uuid.UUID) (int, error)
	GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]*OrderItem, error)
	GetCustomerStats(ctx context.Context, customerID uuid.UUID) (map[string]interface{}, error)
}

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRepository creates a new orders repository
func NewRepository(db *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

// Create creates a new order
func (r *RepositoryImpl) Create(ctx context.Context, order *Order) error {
	query := `
		INSERT INTO orders (id, customer_id, restaurant_id, region_id, delivery_address_id, workflow_state, subtotal, delivery_fee, tax, total, currency_code, scheduled_for, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.Exec(ctx, query,
		order.ID,
		order.CustomerID,
		order.RestaurantID,
		order.RegionID,
		order.DeliveryAddressID,
		order.WorkflowState,
		order.Subtotal,
		order.DeliveryFee,
		order.Tax,
		order.Total,
		order.CurrencyCode,
		order.ScheduledFor,
		order.CreatedAt,
	)

	return err
}

// CreateWithItems creates a new order with its items in a transaction
func (r *RepositoryImpl) CreateWithItems(ctx context.Context, order *Order, items []*OrderItem) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert order
	orderQuery := `
		INSERT INTO orders (id, customer_id, restaurant_id, region_id, delivery_address_id, workflow_state, subtotal, delivery_fee, tax, total, currency_code, scheduled_for, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err = tx.Exec(ctx, orderQuery,
		order.ID,
		order.CustomerID,
		order.RestaurantID,
		order.RegionID,
		order.DeliveryAddressID,
		order.WorkflowState,
		order.Subtotal,
		order.DeliveryFee,
		order.Tax,
		order.Total,
		order.CurrencyCode,
		order.ScheduledFor,
		order.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	// Insert order items
	for _, item := range items {
		itemQuery := `
			INSERT INTO order_items (id, order_id, menu_item_id, quantity, unit_price)
			VALUES ($1, $2, $3, $4, $5)
		`
		_, err = tx.Exec(ctx, itemQuery,
			item.ID,
			order.ID,
			item.MenuItemID,
			item.Quantity,
			item.UnitPrice,
		)
		if err != nil {
			return fmt.Errorf("failed to create order item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves an order by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*Order, error) {
	query := `
		SELECT id, customer_id, restaurant_id, region_id, delivery_address_id, workflow_state, subtotal, delivery_fee, tax, total, currency_code, scheduled_for, created_at
		FROM orders
		WHERE id = $1 AND deleted_at IS NULL
	`

	var order Order
	err := r.db.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.RestaurantID,
		&order.RegionID,
		&order.DeliveryAddressID,
		&order.WorkflowState,
		&order.Subtotal,
		&order.DeliveryFee,
		&order.Tax,
		&order.Total,
		&order.CurrencyCode,
		&order.ScheduledFor,
		&order.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &order, nil
}

// ListByCustomer lists orders for a customer with pagination
func (r *RepositoryImpl) ListByCustomer(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]*Order, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, customer_id, restaurant_id, region_id, delivery_address_id, workflow_state, subtotal, delivery_fee, tax, total, currency_code, scheduled_for, created_at
		FROM orders
		WHERE customer_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, customerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var order Order
		err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.RestaurantID,
			&order.RegionID,
			&order.DeliveryAddressID,
			&order.WorkflowState,
			&order.Subtotal,
			&order.DeliveryFee,
			&order.Tax,
			&order.Total,
			&order.CurrencyCode,
			&order.ScheduledFor,
			&order.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

// ListByCustomerWithItems lists orders for a customer with items in a single query to fix N+1 problem
func (r *RepositoryImpl) ListByCustomerWithItems(ctx context.Context, customerID uuid.UUID, limit, offset int) (map[uuid.UUID][]*OrderItem, []*Order, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT o.id, o.customer_id, o.restaurant_id, o.region_id, o.delivery_address_id, o.workflow_state, 
		       o.subtotal, o.delivery_fee, o.tax, o.total, o.currency_code, o.scheduled_for, o.created_at,
		       oi.id as item_id, oi.order_id as item_order_id, oi.menu_item_id, oi.quantity, oi.unit_price
		FROM orders o
		LEFT JOIN order_items oi ON o.id = oi.order_id
		WHERE o.customer_id = $1 AND o.deleted_at IS NULL
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, customerID, limit, offset)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	ordersMap := make(map[uuid.UUID]*Order)
	itemsMap := make(map[uuid.UUID][]*OrderItem)

	for rows.Next() {
		var order Order
		var itemID, itemOrderID, menuItemID uuid.UUID
		var quantity, unitPrice int
		var itemIDNull, itemOrderIDNull, menuItemIDNull bool

		err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.RestaurantID,
			&order.RegionID,
			&order.DeliveryAddressID,
			&order.WorkflowState,
			&order.Subtotal,
			&order.DeliveryFee,
			&order.Tax,
			&order.Total,
			&order.CurrencyCode,
			&order.ScheduledFor,
			&order.CreatedAt,
			&itemID,
			&itemOrderID,
			&menuItemID,
			&quantity,
			&unitPrice,
		)
		if err != nil {
			return nil, nil, err
		}

		// Check if item fields are NULL (no items for this order)
		itemIDNull = itemID == uuid.Nil
		itemOrderIDNull = itemOrderID == uuid.Nil
		menuItemIDNull = menuItemID == uuid.Nil

		// Store order if not already in map
		if _, exists := ordersMap[order.ID]; !exists {
			ordersMap[order.ID] = &order
		}

		// Add item if it exists
		if !itemIDNull && !itemOrderIDNull && !menuItemIDNull {
			item := &OrderItem{
				ID:         itemID,
				OrderID:    itemOrderID,
				MenuItemID: menuItemID,
				Quantity:   quantity,
				UnitPrice:  unitPrice,
			}
			itemsMap[order.ID] = append(itemsMap[order.ID], item)
		}
	}

	// Convert orders map to slice
	orders := make([]*Order, 0, len(ordersMap))
	for _, order := range ordersMap {
		orders = append(orders, order)
	}

	return itemsMap, orders, nil
}

// ListByRestaurant lists orders for a restaurant with pagination
func (r *RepositoryImpl) ListByRestaurant(ctx context.Context, restaurantID uuid.UUID, limit, offset int) ([]*Order, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, customer_id, restaurant_id, region_id, delivery_address_id, workflow_state, subtotal, delivery_fee, tax, total, currency_code, scheduled_for, created_at
		FROM orders
		WHERE restaurant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, restaurantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var order Order
		err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.RestaurantID,
			&order.RegionID,
			&order.DeliveryAddressID,
			&order.WorkflowState,
			&order.Subtotal,
			&order.DeliveryFee,
			&order.Tax,
			&order.Total,
			&order.CurrencyCode,
			&order.ScheduledFor,
			&order.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

// UpdateWorkflowState updates an order's workflow state
func (r *RepositoryImpl) UpdateWorkflowState(ctx context.Context, id uuid.UUID, state string) error {
	query := `
		UPDATE orders
		SET workflow_state = $2, updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, state)
	return err
}

// GetRestaurantRegion retrieves the region ID for a restaurant
func (r *RepositoryImpl) GetRestaurantRegion(ctx context.Context, restaurantID uuid.UUID) (uuid.UUID, error) {
	query := `
		SELECT region_id
		FROM restaurants
		WHERE id = $1 AND deleted_at IS NULL
	`

	var regionID uuid.UUID
	err := r.db.QueryRow(ctx, query, restaurantID).Scan(&regionID)
	if err != nil {
		return uuid.Nil, err
	}

	return regionID, nil
}

// GetRestaurantStatus retrieves restaurant status information
func (r *RepositoryImpl) GetRestaurantStatus(ctx context.Context, restaurantID uuid.UUID) (bool, bool, string, error) {
	query := `
		SELECT is_active, is_approved, currency_code
		FROM restaurants
		WHERE id = $1 AND deleted_at IS NULL
	`

	var isActive, isApproved bool
	var currencyCode string
	err := r.db.QueryRow(ctx, query, restaurantID).Scan(&isActive, &isApproved, &currencyCode)
	if err != nil {
		return false, false, "", err
	}

	return isActive, isApproved, currencyCode, nil
}

// ValidateAddressOwnership checks if an address belongs to a customer
func (r *RepositoryImpl) ValidateAddressOwnership(ctx context.Context, addressID uuid.UUID, customerID uuid.UUID) (bool, error) {
	query := `
		SELECT 1
		FROM user_addresses
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var exists int
	err := r.db.QueryRow(ctx, query, addressID, customerID).Scan(&exists)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetMenuItemPrice retrieves the price of a menu item
func (r *RepositoryImpl) GetMenuItemPrice(ctx context.Context, menuItemID uuid.UUID) (int, error) {
	query := `
		SELECT price_amount
		FROM menu_items
		WHERE id = $1 AND deleted_at IS NULL
	`

	var price int
	err := r.db.QueryRow(ctx, query, menuItemID).Scan(&price)
	if err != nil {
		return 0, err
	}

	return price, nil
}

// GetOrderItems retrieves items for an order
func (r *RepositoryImpl) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]*OrderItem, error) {
	query := `
		SELECT id, order_id, menu_item_id, quantity, unit_price
		FROM order_items
		WHERE order_id = $1
	`

	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*OrderItem
	for rows.Next() {
		var item OrderItem
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.MenuItemID,
			&item.Quantity,
			&item.UnitPrice,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, nil
}

// GetCustomerStats retrieves order statistics for a customer
func (r *RepositoryImpl) GetCustomerStats(ctx context.Context, customerID uuid.UUID) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_orders,
			COALESCE(SUM(total), 0) as total_spent,
			COUNT(CASE WHEN workflow_state = 'delivered' THEN 1 END) as delivered_orders,
			COUNT(CASE WHEN workflow_state = 'cancelled' THEN 1 END) as cancelled_orders
		FROM orders
		WHERE customer_id = $1 AND deleted_at IS NULL
	`

	var totalOrders, deliveredOrders, cancelledOrders int
	var totalSpent int64
	err := r.db.QueryRow(ctx, query, customerID).Scan(
		&totalOrders,
		&totalSpent,
		&deliveredOrders,
		&cancelledOrders,
	)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_orders":     totalOrders,
		"total_spent":      totalSpent,
		"delivered_orders": deliveredOrders,
		"cancelled_orders": cancelledOrders,
	}, nil
}
