# Testing

## Overview

This document describes the testing strategy, test types, how to run tests, and what critical flows are covered by tests for the Food Delivery API Backend.

## Test Strategy

### Test Pyramid

```
           E2E Tests (5%)
          /            \
     Integration Tests (25%)
    /                    \
   Unit Tests (70%)
```

### Test Types

1. **Unit Tests**: Test individual functions and methods in isolation
2. **Integration Tests**: Test interactions between components with real database
3. **End-to-End Tests**: Test complete user flows through HTTP API

### Testing Philosophy

- Test business logic, not implementation details
- Use table-driven tests for multiple scenarios
- Mock external dependencies (payment provider, email service)
- Use test database for integration tests
- Tests should be fast and deterministic
- Aim for high coverage of critical paths

## How to Run Tests

### Run All Tests

```bash
make test
```

### Run Unit Tests Only

```bash
go test ./... -short
```

### Run Integration Tests

```bash
go test ./... -v
```

### Run Specific Test Package

```bash
go test ./internal/orders/...
```

### Run Specific Test

```bash
go test ./internal/orders/ -run TestCreateOrder
```

### Run with Coverage

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run with Race Detection

```bash
go test ./... -race
```

## Critical Flows Covered

### Authentication Tests

- User registration with valid data
- User registration with invalid email
- User registration with weak password
- User login with correct credentials
- User login with incorrect credentials
- Token refresh with valid token
- Token refresh with expired token
- Token refresh with revoked token
- Logout invalidates refresh token

**Example**:
```go
func TestAuthService_Register(t *testing.T) {
    tests := []struct {
        name    string
        req     RegisterRequest
        wantErr error
    }{
        {
            name: "valid registration",
            req: RegisterRequest{
                Email:    "test@example.com",
                Password: "SecurePass123!",
                Role:     "customer",
            },
            wantErr: nil,
        },
        {
            name: "invalid email",
            req: RegisterRequest{
                Email:    "invalid-email",
                Password: "SecurePass123!",
                Role:     "customer",
            },
            wantErr: ErrInvalidEmail,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := service.Register(context.Background(), tt.req)
            assert.Equal(t, tt.wantErr, err)
        })
    }
}
```

### Order Creation Tests

- Create ASAP order with valid data
- Create scheduled order with valid future time
- Create scheduled order with past time (should fail)
- Create order with inactive restaurant (should fail)
- Create order with unavailable menu item (should fail)
- Server-side price calculation
- Server-side tax calculation
- Order creation with idempotency key
- Duplicate order creation with same idempotency key

**Example**:
```go
func TestOrderService_CreateOrder(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    restaurant := createTestRestaurant(t)
    menuItem := createTestMenuItem(t, restaurant.ID)
    customer := createTestCustomer(t)
    
    t.Run("create ASAP order", func(t *testing.T) {
        req := CreateOrderRequest{
            RestaurantID: restaurant.ID,
            DeliveryType: "asap",
            Items: []OrderItemRequest{
                {MenuItemID: menuItem.ID, Quantity: 2},
            },
        }
        
        order, err := orderService.CreateOrder(context.Background(), customer.ID, req)
        
        assert.NoError(t, err)
        assert.Equal(t, "pending_payment", order.Status)
        assert.Equal(t, 2000, order.SubtotalAmount) // 2 * 1000
    })
    
    t.Run("server-side pricing", func(t *testing.T) {
        req := CreateOrderRequest{
            RestaurantID: restaurant.ID,
            DeliveryType: "asap",
            Items: []OrderItemRequest{
                {MenuItemID: menuItem.ID, Quantity: 2},
            },
            // Try to send wrong price from client
            SubtotalAmount: 100, // Wrong!
        }
        
        order, err := orderService.CreateOrder(context.Background(), customer.ID, req)
        
        assert.NoError(t, err)
        assert.NotEqual(t, 100, order.SubtotalAmount) // Server ignored client price
        assert.Equal(t, 2000, order.SubtotalAmount) // Server calculated correctly
    })
}
```

### Role-Based Authorization Tests

- Customer cannot access admin endpoints
- Restaurant owner cannot access another owner's restaurants
- Courier cannot access customer orders
- Admin can access all resources
- Ownership checks enforced on all CRUD operations

**Example**:
```go
func TestAuthorization_RoleBasedAccess(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    customer := createTestUser(t, "customer")
    admin := createTestUser(t, "admin")
    
    t.Run("customer cannot access admin endpoint", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/v1/admin/users", nil)
        req = req.WithContext(contextWithUser(req.Context(), customer))
        
        resp := executeRequest(req)
        
        assert.Equal(t, http.StatusForbidden, resp.Code)
    })
    
    t.Run("admin can access admin endpoint", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/v1/admin/users", nil)
        req = req.WithContext(contextWithUser(req.Context(), admin))
        
        resp := executeRequest(req)
        
        assert.Equal(t, http.StatusOK, resp.Code)
    })
}
```

### Delivery Claim Concurrency Tests

- Multiple couriers cannot claim same delivery
- SKIP LOCKED prevents double-claim
- Delivery assignment is atomic

**Example**:
```go
func TestDeliveryService_ConcurrentClaim(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    delivery := createTestDelivery(t)
    courier1 := createTestCourier(t)
    courier2 := createTestCourier(t)
    
    var wg sync.WaitGroup
    var successCount int32
    var mu sync.Mutex
    
    // Both couriers try to claim simultaneously
    for _, courier := range []*User{courier1, courier2} {
        wg.Add(1)
        go func(c *User) {
            defer wg.Done()
            
            err := deliveryService.Claim(context.Background(), delivery.ID, c.ID)
            if err == nil {
                mu.Lock()
                successCount++
                mu.Unlock()
            }
        }(courier)
    }
    
    wg.Wait()
    
    // Only one should succeed
    assert.Equal(t, int32(1), successCount)
}
```

### Scheduled Order Activation Tests

- Scheduled order activates when due
- Scheduled order does not activate before due time
- Worker validates restaurant hours before activation
- Scheduled order cancelled if restaurant closed at activation time

**Example**:
```go
func TestScheduledOrderActivation(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    restaurant := createTestRestaurant(t)
    setRestaurantHours(t, restaurant.ID, "09:00", "21:00")
    
    t.Run("activates when due", func(t *testing.T) {
        scheduledFor := time.Now().Add(1 * time.Hour)
        order := createScheduledOrder(t, restaurant.ID, scheduledFor)
        
        // Fast-forward time
        clock := NewMockClock(scheduledFor.Add(1 * time.Minute))
        
        err := worker.ActivateScheduledOrder(context.Background(), order.ID, clock)
        
        assert.NoError(t, err)
        
        updatedOrder, _ := orderRepo.GetByID(context.Background(), order.ID)
        assert.Equal(t, "pending", updatedOrder.Status)
        assert.NotNil(t, updatedOrder.ActivatedAt)
    })
    
    t.Run("does not activate before due time", func(t *testing.T) {
        scheduledFor := time.Now().Add(1 * time.Hour)
        order := createScheduledOrder(t, restaurant.ID, scheduledFor)
        
        clock := NewMockClock(scheduledFor.Add(-1 * time.Minute)) // Before due time
        
        err := worker.ActivateScheduledOrder(context.Background(), order.ID, clock)
        
        assert.Error(t, err)
        
        updatedOrder, _ := orderRepo.GetByID(context.Background(), order.ID)
        assert.Equal(t, "scheduled", updatedOrder.Status)
    })
}
```

### Worker Processing Tests

- Outbox event processed successfully
- Worker retries on transient failure
- Worker marks dead-letter after max retries
- Worker processes only committed transactions
- Worker is idempotent

**Example**:
```go
func TestOutboxWorker_ProcessEvent(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    t.Run("processes event successfully", func(t *testing.T) {
        event := createOutboxEvent(t, "order_created", orderID)
        
        err := worker.ProcessEvent(context.Background(), event)
        
        assert.NoError(t, err)
        
        updatedEvent, _ := outboxRepo.GetByID(context.Background(), event.ID)
        assert.True(t, updatedEvent.Processed)
    })
    
    t.Run("retries on transient failure", func(t *testing.T) {
        event := createOutboxEvent(t, "order_created", orderID)
        
        // Mock transient failure
        notificationService.ShouldFailTransiently = true
        
        err := worker.ProcessEvent(context.Background(), event)
        
        assert.Error(t, err)
        
        job, _ := jobRepo.GetByAggregate(context.Background(), "order", orderID)
        assert.Equal(t, 1, job.RetryCount)
    })
}
```

### Invalid State Transition Tests

- Order cannot transition from pending to ready (must go through accepted, preparing)
- Order cannot be cancelled after delivery
- Delivery cannot be completed before pickup
- Invalid transitions return domain errors

**Example**:
```go
func TestOrderWorkflow_InvalidTransitions(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    order := createOrder(t, "pending")
    
    t.Run("cannot skip states", func(t *testing.T) {
        err := orderService.UpdateStatus(context.Background(), order.ID, "ready_for_pickup")
        
        assert.Error(t, err)
        assert.Equal(t, ErrInvalidTransition, err)
    })
    
    t.Run("valid transition sequence", func(t *testing.T) {
        // pending -> accepted
        err := orderService.UpdateStatus(context.Background(), order.ID, "accepted")
        assert.NoError(t, err)
        
        // accepted -> preparing
        err = orderService.UpdateStatus(context.Background(), order.ID, "preparing")
        assert.NoError(t, err)
        
        // preparing -> ready_for_pickup
        err = orderService.UpdateStatus(context.Background(), order.ID, "ready_for_pickup")
        assert.NoError(t, err)
    })
}
```

### Idempotency Tests

- Duplicate request with same idempotency key returns original response
- Idempotency key expires after TTL
- Different scopes allow same key

**Example**:
```go
func TestIdempotencyService(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    t.Run("duplicate request returns original response", func(t *testing.T) {
        key := "test-key-123"
        scope := "user-123"
        
        req := httptest.NewRequest("POST", "/v1/orders", strings.NewReader(`{...}`))
        req.Header.Set("Idempotency-Key", key)
        
        // First request
        resp1 := executeRequest(req)
        assert.Equal(t, http.StatusCreated, resp1.Code)
        
        // Duplicate request
        resp2 := executeRequest(req)
        assert.Equal(t, http.StatusCreated, resp2.Code)
        assert.Equal(t, resp1.Body.String(), resp2.Body.String())
    })
}
```

### Audit Log Tests

- Sensitive admin actions create audit logs
- Audit logs include actor, action, entity, metadata
- Audit logs are immutable

**Example**:
```go
func TestAuditService_LogAction(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    admin := createTestUser(t, "admin")
    restaurant := createTestRestaurant(t)
    
    auditLog := &AuditLog{
        ActorID:    admin.ID,
        ActorRole:  "admin",
        Action:     "restaurant_approved",
        EntityType: "restaurant",
        EntityID:   restaurant.ID,
        Metadata:   map[string]interface{}{"reason": "Met requirements"},
    }
    
    err := auditService.Create(context.Background(), auditLog)
    
    assert.NoError(t, err)
    
    logs, _ := auditRepo.ListByEntity(context.Background(), "restaurant", restaurant.ID)
    assert.Len(t, logs, 1)
    assert.Equal(t, "restaurant_approved", logs[0].Action)
}
```

### Rate Limiting Tests

- Rate limit enforced on protected endpoints
- Rate limit headers returned
- Rate limit resets after window

**Example**:
```go
func TestRateLimitMiddleware(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    middleware := RateLimitMiddleware(5, time.Minute) // 5 requests per minute
    
    t.Run("enforces rate limit", func(t *testing.T) {
        for i := 0; i < 6; i++ {
            req := httptest.NewRequest("POST", "/v1/auth/login", nil)
            req.RemoteAddr = "127.0.0.1"
            
            handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
            }))
            
            recorder := httptest.NewRecorder()
            handler.ServeHTTP(recorder, req)
            
            if i < 5 {
                assert.Equal(t, http.StatusOK, recorder.Code)
            } else {
                assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
            }
        }
    })
}
```

### Tax Calculation Tests

- Tax calculated server-side
- Inclusive tax mode
- Exclusive tax mode
- Delivery fee taxation
- Tax snapshot immutability
- Tax changes don't affect historical orders

**Example**:
```go
func TestTaxService_Calculate(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    taxZone := createTestTaxZone(t)
    taxRate := createTestTaxRate(t, taxZone.ID, 800, false) // 8% exclusive
    
    t.Run("exclusive tax calculation", func(t *testing.T) {
        items := []TaxableItem{
            {Amount: 1000, TaxCategoryID: taxRate.TaxCategoryID},
        }
        
        result, err := taxService.Calculate(context.Background(), taxZone.ID, items, 200)
        
        assert.NoError(t, err)
        assert.Equal(t, 1000, result.SubtotalAmount)
        assert.Equal(t, 80, result.TaxAmount) // 8% of 1000
        assert.Equal(t, 1280, result.TotalAmount)
    })
    
    t.Run("inclusive tax calculation", func(t *testing.T) {
        inclusiveRate := createTestTaxRate(t, taxZone.ID, 800, true)
        items := []TaxableItem{
            {Amount: 1000, TaxCategoryID: inclusiveRate.TaxCategoryID},
        }
        
        result, err := taxService.Calculate(context.Background(), taxZone.ID, items, 200)
        
        assert.NoError(t, err)
        assert.Equal(t, 926, result.SubtotalAmount) // 1000 / 1.08
        assert.Equal(t, 74, result.TaxAmount)
        assert.Equal(t, 1000, result.TotalAmount)
    })
}
```

### Timezone-Aware Scheduling Tests

- Scheduled time validated in restaurant timezone
- Restaurant hours checked in local time
- Worker activation uses UTC but validates local time

**Example**:
```go
func TestScheduledOrder_TimezoneValidation(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    // Restaurant in New York (America/New_York)
    restaurant := createTestRestaurant(t)
    restaurant.Timezone = "America/New_York"
    restaurantRepo.Update(context.Background(), restaurant)
    
    setRestaurantHours(t, restaurant.ID, "09:00", "21:00")
    
    t.Run("valid scheduled time in local hours", func(t *testing.T) {
        // 2 PM New York time
        scheduledFor := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
        
        req := CreateOrderRequest{
            RestaurantID: restaurant.ID,
            DeliveryType: "scheduled",
            ScheduledFor: scheduledFor,
        }
        
        order, err := orderService.CreateOrder(context.Background(), customerID, req)
        
        assert.NoError(t, err)
        assert.Equal(t, "scheduled", order.Status)
    })
    
    t.Run("invalid scheduled time outside local hours", func(t *testing.T) {
        // 2 AM New York time (restaurant closed)
        scheduledFor := time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC) // 2 AM EST
        
        req := CreateOrderRequest{
            RestaurantID: restaurant.ID,
            DeliveryType: "scheduled",
            ScheduledFor: scheduledFor,
        }
        
        _, err := orderService.CreateOrder(context.Background(), customerID, req)
        
        assert.Error(t, err)
        assert.Equal(t, ErrRestaurantClosed, err)
    })
}
```

### Region-Specific Configuration Tests

- Orders use region currency
- Taxes calculated using region tax zone
- Delivery windows use region defaults
- Region config changes affect new orders only

**Example**:
```go
func TestRegionConfiguration(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    region := createTestRegion(t, "USD", "America/New_York")
    regionConfig := createTestRegionConfig(t, region.ID, 1500, 30) // 15% fee, 30 min window
    
    restaurant := createTestRestaurant(t, region.ID)
    
    t.Run("order uses region currency", func(t *testing.T) {
        order := createOrder(t, restaurant.ID)
        
        assert.Equal(t, "USD", order.CurrencyCode)
    })
    
    t.Run("delivery fee uses region config", func(t *testing.T) {
        order := createOrder(t, restaurant.ID)
        
        assert.Equal(t, 30, order.DeliveryWindowMinutes)
    })
}
```

### Customer CRUD Ownership Tests

- Customer can CRUD own addresses
- Customer cannot access other customers' addresses
- Customer cannot access other customers' orders

**Example**:
```go
func TestCustomerOwnership(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    customer1 := createTestCustomer(t)
    customer2 := createTestCustomer(t)
    address := createTestAddress(t, customer1.ID)
    
    t.Run("customer can access own address", func(t *testing.T) {
        ctx := contextWithUser(context.Background(), customer1)
        addr, err := addressService.GetByID(ctx, address.ID)
        
        assert.NoError(t, err)
        assert.Equal(t, address.ID, addr.ID)
    })
    
    t.Run("customer cannot access other customer's address", func(t *testing.T) {
        ctx := contextWithUser(context.Background(), customer2)
        _, err := addressService.GetByID(ctx, address.ID)
        
        assert.Error(t, err)
        assert.Equal(t, ErrForbidden, err)
    })
}
```

### Restaurant Owner Ownership Tests

- Restaurant owner can CRUD own restaurants
- Restaurant owner cannot access another owner's restaurants
- Restaurant owner can manage own menu items

**Example**:
```go
func TestRestaurantOwnerOwnership(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    owner1 := createTestUser(t, "restaurant_owner")
    owner2 := createTestUser(t, "restaurant_owner")
    restaurant1 := createTestRestaurant(t, owner1.ID)
    restaurant2 := createTestRestaurant(t, owner2.ID)
    
    t.Run("owner can access own restaurant", func(t *testing.T) {
        ctx := contextWithUser(context.Background(), owner1)
        restaurant, err := restaurantService.GetByID(ctx, restaurant1.ID)
        
        assert.NoError(t, err)
        assert.Equal(t, restaurant1.ID, restaurant.ID)
    })
    
    t.Run("owner cannot access another owner's restaurant", func(t *testing.T) {
        ctx := contextWithUser(context.Background(), owner1)
        _, err := restaurantService.GetByID(ctx, restaurant2.ID)
        
        assert.Error(t, err)
        assert.Equal(t, ErrForbidden, err)
    })
}
```

### Courier Access Control Tests

- Courier can access available deliveries
- Courier can claim delivery
- Courier cannot access admin endpoints
- Courier cannot access restaurant owner endpoints

**Example**:
```go
func TestCourierAccessControl(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    courier := createTestUser(t, "courier")
    delivery := createTestDelivery(t)
    
    t.Run("courier can claim delivery", func(t *testing.T) {
        ctx := contextWithUser(context.Background(), courier)
        err := deliveryService.Claim(ctx, delivery.ID, courier.ID)
        
        assert.NoError(t, err)
    })
    
    t.Run("courier cannot access admin endpoint", func(t *testing.T) {
        ctx := contextWithUser(context.Background(), courier)
        req := httptest.NewRequest("GET", "/v1/admin/users", nil)
        req = req.WithContext(ctx)
        
        resp := executeRequest(req)
        
        assert.Equal(t, http.StatusForbidden, resp.Code)
    })
}
```

### Admin CRUD Tests

- Admin can CRUD regions
- Admin can CRUD tax configuration
- Admin can suspend users
- Admin actions create audit logs

**Example**:
```go
func TestAdminCRUD(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    admin := createTestUser(t, "admin")
    
    t.Run("admin can create region", func(t *testing.T) {
        ctx := contextWithUser(context.Background(), admin)
        req := CreateRegionRequest{
            Code: "US-NY",
            Name: "New York",
            CountryCode: "US",
            Timezone: "America/New_York",
            CurrencyCode: "USD",
        }
        
        region, err := adminService.CreateRegion(ctx, req)
        
        assert.NoError(t, err)
        assert.Equal(t, "US-NY", region.Code)
    })
    
    t.Run("admin action creates audit log", func(t *testing.T) {
        ctx := contextWithUser(context.Background(), admin)
        
        user := createTestUser(t, "customer")
        err := adminService.SuspendUser(ctx, user.ID, "Policy violation")
        
        assert.NoError(t, err)
        
        logs, _ := auditRepo.ListByActor(ctx, admin.ID)
        assert.Len(t, logs, 1)
        assert.Equal(t, "user_suspended", logs[0].Action)
    })
}
```

### Soft Delete Tests

- Soft delete excludes records from default listings
- Restore endpoint brings back soft-deleted records
- Historically referenced resources not hard deleted

**Example**:
```go
func TestSoftDelete(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    restaurant := createTestRestaurant(t)
    
    t.Run("soft delete excludes from listings", func(t *testing.T) {
        err := restaurantService.Delete(context.Background(), restaurant.ID)
        assert.NoError(t, err)
        
        restaurants, _ := restaurantService.List(context.Background(), ListRestaurantsRequest{})
        
        // Restaurant should not be in list
        for _, r := range restaurants {
            assert.NotEqual(t, restaurant.ID, r.ID)
        }
    })
    
    t.Run("restore brings back restaurant", func(t *testing.T) {
        err := restaurantService.Restore(context.Background(), restaurant.ID)
        assert.NoError(t, err)
        
        restaurants, _ := restaurantService.List(context.Background(), ListRestaurantsRequest{})
        
        // Restaurant should be in list
        found := false
        for _, r := range restaurants {
            if r.ID == restaurant.ID {
                found = true
                break
            }
        }
        assert.True(t, found)
    })
}
```

### Historically Referenced Resources Tests

- Menu items in orders cannot be hard deleted
- Tax rates used by orders can only be deactivated
- Orders cannot be hard deleted

**Example**:
```go
func TestHistoricallyReferencedResources(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    menuItem := createTestMenuItem(t)
    order := createOrderWithItem(t, menuItem.ID)
    
    t.Run("cannot hard delete menu item in orders", func(t *testing.T) {
        err := menuItemService.HardDelete(context.Background(), menuItem.ID)
        
        assert.Error(t, err)
        assert.Equal(t, ErrReferencedByOrders, err)
    })
    
    t.Run("can deactivate menu item", func(t *testing.T) {
        err := menuItemService.Deactivate(context.Background(), menuItem.ID)
        
        assert.NoError(t, err)
        
        menuItem, _ := menuItemRepo.GetByID(context.Background(), menuItem.ID)
        assert.False(t, menuItem.IsActive)
    })
}
```

### CRUD Endpoint Enforcement Tests

- List endpoints enforce region scope
- List endpoints enforce ownership scope
- Update endpoints enforce workflow state
- Delete endpoints enforce soft delete rules

**Example**:
```go
func TestCRUDEndpointEnforcement(t *testing.T) {
    setupTestDatabase(t)
    defer cleanupTestDatabase(t)
    
    customer := createTestCustomer(t)
    order := createOrder(t, customer.ID)
    
    t.Run("customer cannot update order status", func(t *testing.T) {
        ctx := contextWithUser(context.Background(), customer)
        err := orderService.UpdateStatus(ctx, order.ID, "accepted")
        
        assert.Error(t, err)
        assert.Equal(t, ErrForbidden, err)
    })
    
    t.Run("customer cannot hard delete order", func(t *testing.T) {
        ctx := contextWithUser(context.Background(), customer)
        err := orderService.HardDelete(ctx, order.ID)
        
        assert.Error(t, err)
        assert.Equal(t, ErrForbidden, err)
    })
}
```

## Test Utilities

### Test Database Setup

```go
func setupTestDatabase(t *testing.T) {
    t.Helper()
    
    // Create test database
    db := createTestDBConnection(t)
    
    // Run migrations
    runMigrations(t, db)
    
    // Seed test data
    seedTestData(t, db)
}

func cleanupTestDatabase(t *testing.T) {
    t.Helper()
    
    // Drop test database
    dropTestDB(t)
}
```

### Mock Clock

```go
type MockClock struct {
    now time.Time
}

func NewMockClock(now time.Time) *MockClock {
    return &MockClock{now: now}
}

func (m *MockClock) Now() time.Time {
    return m.now
}

func (m *MockClock) Set(now time.Time) {
    m.now = now
}
```

### Mock Payment Provider

```go
type MockPaymentProvider struct {
    ShouldSucceed bool
}

func (m *MockPaymentProvider) Charge(ctx context.Context, token string, amount int, currency string) (PaymentResult, error) {
    if m.ShouldSucceed {
        return PaymentResult{Status: "succeeded"}, nil
    }
    return PaymentResult{Status: "failed"}, errors.New("payment failed")
}
```

### Test Helpers

```go
func createTestUser(t *testing.T, role string) *User {
    t.Helper()
    
    user := &User{
        Email:    fmt.Sprintf("%s@example.com", uuid.New().String()),
        Password: hashPassword(t, "SecurePass123!"),
        Role:     role,
    }
    
    err := userRepo.Create(context.Background(), user)
    require.NoError(t, err)
    
    return user
}

func createTestRestaurant(t *testing.T) *Restaurant {
    t.Helper()
    
    restaurant := &Restaurant{
        OwnerID:  createTestUser(t, "restaurant_owner").ID,
        Name:     "Test Restaurant",
        Currency: "USD",
    }
    
    err := restaurantRepo.Create(context.Background(), restaurant)
    require.NoError(t, err)
    
    return restaurant
}

func createTestOrder(t *testing.T, restaurantID uuid.UUID) *Order {
    t.Helper()
    
    order := &Order{
        RestaurantID:    restaurantID,
        CustomerID:      createTestUser(t, "customer").ID,
        Status:         "pending",
        CurrencyCode:   "USD",
        SubtotalAmount: 1000,
        TaxAmount:      80,
        TotalAmount:    1080,
    }
    
    err := orderRepo.Create(context.Background(), order)
    require.NoError(t, err)
    
    return order
}
```

## Test Coverage Goals

- Overall coverage: 80%+
- Critical business logic: 95%+
- Authentication/authorization: 90%+
- Order workflow: 95%+
- Tax calculation: 95%+
- Worker processing: 90%+

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: hopper_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      - name: Run tests
        run: go test ./... -v -coverprofile=coverage.out
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

## Summary

Testing includes:
- Unit tests for services and business logic
- Integration tests for API endpoints with real database
- Tests for authentication and authorization
- Tests for order creation logic
- Tests for role-based access control
- Tests for delivery claim concurrency
- Tests for scheduled order activation
- Tests for worker processing
- Tests for invalid state transitions
- Tests for idempotency
- Tests for audit log creation
- Tests for rate limiting
- Tests for tax calculation
- Tests for timezone-aware scheduling
- Tests for region-specific configuration
- Tests for customer CRUD ownership
- Tests for restaurant owner CRUD ownership
- Tests for courier access control
- Tests for admin CRUD
- Tests for soft delete behavior
- Tests for historically referenced resources
- Tests for CRUD endpoint enforcement
