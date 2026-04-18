# API Documentation

## Overview

This document describes the REST API endpoints, authentication model, roles and permissions, workflow state transitions, and CRUD behavior by role for the Food Delivery API Backend.

## Endpoint Summary

### Authentication
- `POST /v1/auth/register` - Register new user
- `POST /v1/auth/login` - Login user
- `POST /v1/auth/refresh` - Refresh access token
- `POST /v1/auth/logout` - Logout user (invalidate refresh token)
- `GET /v1/me` - Get current user profile
- `PATCH /v1/me` - Update current user profile

### Customer Profile / Addresses
- `GET /v1/me/addresses` - List customer addresses
- `POST /v1/me/addresses` - Create address
- `GET /v1/me/addresses/{id}` - Get address
- `PATCH /v1/me/addresses/{id}` - Update address
- `DELETE /v1/me/addresses/{id}` - Delete address

### Customer Discovery
- `GET /v1/restaurants` - List restaurants (filtered by region)
- `GET /v1/restaurants/{id}` - Get restaurant details
- `GET /v1/restaurants/{id}/menu` - Get restaurant menu

### Customer Orders
- `POST /v1/orders` - Create order
- `GET /v1/orders/{id}` - Get order details
- `GET /v1/me/orders` - List customer orders
- `PATCH /v1/orders/{id}/cancel` - Cancel order

### Customer Payments
- `GET /v1/me/payments` - List customer payments
- `GET /v1/me/payments/{id}` - Get payment details

### Restaurant Owner Restaurants
- `GET /v1/owner/restaurants` - List owner restaurants
- `POST /v1/owner/restaurants` - Create restaurant
- `GET /v1/owner/restaurants/{id}` - Get restaurant
- `PATCH /v1/owner/restaurants/{id}` - Update restaurant
- `DELETE /v1/owner/restaurants/{id}` - Soft delete restaurant

### Restaurant Owner Hours
- `GET /v1/owner/restaurants/{id}/hours` - List restaurant hours
- `POST /v1/owner/restaurants/{id}/hours` - Create hours
- `GET /v1/owner/restaurants/{id}/hours/{hourId}` - Get hours
- `PATCH /v1/owner/restaurants/{id}/hours/{hourId}` - Update hours
- `DELETE /v1/owner/restaurants/{id}/hours/{hourId}` - Delete hours

### Restaurant Owner Menu Items
- `GET /v1/owner/restaurants/{id}/menu-items` - List menu items
- `POST /v1/owner/restaurants/{id}/menu-items` - Create menu item
- `GET /v1/owner/menu-items/{id}` - Get menu item
- `PATCH /v1/owner/menu-items/{id}` - Update menu item
- `DELETE /v1/owner/menu-items/{id}` - Soft delete menu item

### Restaurant Owner Order Management
- `GET /v1/owner/orders` - List restaurant orders
- `GET /v1/owner/orders/{id}` - Get order details
- `PATCH /v1/owner/orders/{id}/accept` - Accept order
- `PATCH /v1/owner/orders/{id}/reject` - Reject order
- `PATCH /v1/owner/orders/{id}/preparing` - Mark order preparing
- `PATCH /v1/owner/orders/{id}/ready` - Mark order ready

### Courier
- `GET /v1/courier/profile` - Get courier profile
- `PATCH /v1/courier/profile` - Update courier profile
- `GET /v1/courier/deliveries/available` - List available deliveries
- `GET /v1/courier/deliveries` - List courier deliveries
- `GET /v1/courier/deliveries/{id}` - Get delivery details
- `POST /v1/courier/deliveries/{id}/claim` - Claim delivery
- `PATCH /v1/courier/deliveries/{id}/pickup` - Mark pickup
- `PATCH /v1/courier/deliveries/{id}/complete` - Mark complete

### Admin Users
- `GET /v1/admin/users` - List users
- `POST /v1/admin/users` - Create user
- `GET /v1/admin/users/{id}` - Get user
- `PATCH /v1/admin/users/{id}` - Update user
- `DELETE /v1/admin/users/{id}` - Soft delete user
- `PATCH /v1/admin/users/{id}/suspend` - Suspend user
- `PATCH /v1/admin/users/{id}/restore` - Restore user

### Admin Restaurants
- `GET /v1/admin/restaurants` - List restaurants
- `GET /v1/admin/restaurants/pending` - List pending restaurants
- `GET /v1/admin/restaurants/{id}` - Get restaurant
- `PATCH /v1/admin/restaurants/{id}` - Update restaurant
- `PATCH /v1/admin/restaurants/{id}/approve` - Approve restaurant
- `DELETE /v1/admin/restaurants/{id}` - Archive restaurant
- `PATCH /v1/admin/restaurants/{id}/restore` - Restore restaurant

### Admin Orders
- `GET /v1/admin/orders` - List orders
- `GET /v1/admin/orders/{id}` - Get order
- `PATCH /v1/admin/orders/{id}` - Update order (with audit)
- `PATCH /v1/admin/orders/{id}/cancel` - Cancel order (with audit)

### Admin Payments
- `GET /v1/admin/payments` - List payments
- `GET /v1/admin/payments/{id}` - Get payment
- `PATCH /v1/admin/payments/{id}` - Update payment status (with audit)

### Admin Regions
- `GET /v1/admin/regions` - List regions
- `POST /v1/admin/regions` - Create region
- `GET /v1/admin/regions/{id}` - Get region
- `PATCH /v1/admin/regions/{id}` - Update region
- `DELETE /v1/admin/regions/{id}` - Deactivate region
- `PATCH /v1/admin/regions/{id}/restore` - Restore region

### Admin Region Configs
- `GET /v1/admin/region-configs` - List region configs
- `POST /v1/admin/region-configs` - Create region config
- `GET /v1/admin/region-configs/{id}` - Get region config
- `PATCH /v1/admin/region-configs/{id}` - Update region config
- `DELETE /v1/admin/region-configs/{id}` - Delete region config

### Admin Tax Categories
- `GET /v1/admin/tax-categories` - List tax categories
- `POST /v1/admin/tax-categories` - Create tax category
- `GET /v1/admin/tax-categories/{id}` - Get tax category
- `PATCH /v1/admin/tax-categories/{id}` - Update tax category
- `DELETE /v1/admin/tax-categories/{id}` - Delete tax category

### Admin Tax Zones
- `GET /v1/admin/tax-zones` - List tax zones
- `POST /v1/admin/tax-zones` - Create tax zone
- `GET /v1/admin/tax-zones/{id}` - Get tax zone
- `PATCH /v1/admin/tax-zones/{id}` - Update tax zone
- `DELETE /v1/admin/tax-zones/{id}` - Delete tax zone

### Admin Tax Rates
- `GET /v1/admin/tax-rates` - List tax rates
- `POST /v1/admin/tax-rates` - Create tax rate
- `GET /v1/admin/tax-rates/{id}` - Get tax rate
- `PATCH /v1/admin/tax-rates/{id}` - Update tax rate
- `DELETE /v1/admin/tax-rates/{id}` - Deactivate tax rate

### Admin Audit / Inspection
- `GET /v1/admin/audit-logs` - List audit logs
- `GET /v1/admin/audit-logs/{id}` - Get audit log

### System
- `GET /healthz` - Health check
- `GET /readyz` - Readiness check
- `GET /metrics` - Metrics endpoint (Prometheus format)

## Authentication Model

### JWT Access Tokens

- Short-lived access tokens (15 minutes TTL)
- Contains user ID and role in claims
- Required for all protected endpoints
- Sent in `Authorization: Bearer <token>` header

### Refresh Tokens

- Long-lived refresh tokens (7-30 days TTL)
- Stored in database (`refresh_tokens` table)
- Rotated on every use
- Can be revoked (logout)
- Used to obtain new access tokens

### Token Structure

Access token claims:
```json
{
  "sub": "user_id",
  "role": "customer|restaurant_owner|courier|admin",
  "iat": 1234567890,
  "exp": 1234567890
}
```

### Auth Flow

1. **Register**: User provides email, password, role → creates user, returns access + refresh tokens
2. **Login**: User provides email, password → validates, returns access + refresh tokens
3. **Refresh**: User provides refresh token → validates, returns new access + refresh tokens
4. **Logout**: User provides refresh token → marks as revoked

## Roles and Permissions

### Customer

**Permissions**:
- Read/update own profile
- Full CRUD on own addresses
- Read restaurants and menus
- Create orders, read own orders, cancel own orders (when allowed)
- Read own payment status

**Restrictions**:
- Cannot access admin/owner/courier resources
- Cannot access other customers' data
- Cannot modify orders after certain states

### Restaurant Owner

**Permissions**:
- Create/read/update own restaurants (soft delete)
- Full CRUD on own restaurant hours
- Full CRUD on own menu items (activate/deactivate, not hard delete if referenced)
- Read/list and workflow updates for own orders
- Accept/reject/preparing/ready for own orders

**Restrictions**:
- Cannot manage restaurants owned by others
- Cannot delete historical orders
- Cannot access customer/courier/admin resources

### Courier

**Permissions**:
- Read/update limited courier profile
- Read available deliveries
- Claim deliveries
- Read own assigned deliveries
- Update pickup/complete states

**Restrictions**:
- Cannot delete delivery history
- Cannot access customer/restaurant admin CRUD
- Cannot access admin resources

### Admin

**Permissions**:
- Full administrative read/list on users (limited update, controlled disable/suspend)
- Full admin read/list/approve/update/suspend/archive on restaurants
- Full CRUD on regions and region configs
- Full CRUD on tax categories, tax zones, tax rates
- Read-only on audit logs
- Read/list and limited administrative overrides on orders/payments (with audit)

**Restrictions**:
- Sensitive actions must be audited
- Avoid destructive hard delete unless justified
- Overrides create audit logs

## Workflow/State Transition Summary

### Order Statuses

| Status | Description | Allowed Next States |
|--------|-------------|---------------------|
| pending_payment | Order created, awaiting payment | scheduled, pending |
| scheduled | Scheduled order, awaiting activation | pending |
| pending | Order placed, awaiting restaurant acceptance | accepted, rejected |
| accepted | Restaurant accepted, awaiting preparation | preparing, cancelled |
| rejected | Restaurant rejected | (terminal) |
| preparing | Restaurant preparing | ready_for_pickup, cancelled |
| ready_for_pickup | Order ready for pickup | picked_up, cancelled |
| picked_up | Courier picked up | delivered, failed |
| delivered | Order delivered | (terminal) |
| cancelled | Order cancelled | (terminal) |

### Delivery Statuses

| Status | Description | Allowed Next States |
|--------|-------------|---------------------|
| pending_schedule | Delivery scheduled, awaiting order activation | unassigned |
| unassigned | Delivery unassigned, awaiting courier claim | assigned |
| assigned | Courier assigned, awaiting pickup | picked_up, cancelled |
| picked_up | Courier picked up, awaiting delivery | delivered, failed |
| delivered | Delivery completed | (terminal) |
| failed | Delivery failed | (terminal) |
| cancelled | Delivery cancelled | (terminal) |

### Transition Rules

- Order cannot be marked preparing before accepted
- Order cannot be marked ready before preparing
- Courier cannot pick up before claiming/assignment
- Courier cannot complete delivery before pickup
- Scheduled orders must not enter active fulfillment before activation window
- Invalid transitions return domain errors

## Scheduled Order Behavior

### ASAP vs Scheduled

**ASAP Orders**:
- `delivery_type = "asap"`
- `scheduled_for = null`
- Order moves directly to `pending` after payment
- Restaurant can accept immediately

**Scheduled Orders**:
- `delivery_type = "scheduled"`
- `scheduled_for` = future timestamp (in restaurant timezone)
- Order moves to `scheduled` after payment
- Worker activates order when `scheduled_for <= now()`
- Restaurant can only accept after activation

### Activation Flow

1. Customer creates scheduled order with `scheduled_for` in future
2. Payment completes → order status = `scheduled`
3. Worker polls for scheduled orders where `scheduled_for <= now()`
4. Worker validates restaurant is still open
5. Worker activates order → status = `pending`, sets `activated_at`
6. Restaurant can now see and accept order

### Validation Rules

- `scheduled_for` must be in the future
- `scheduled_for` must be within restaurant operating hours
- `scheduled_for` must be within region business hours
- Delivery window must be reasonable (configurable per region)
- Cancellation rules differ before vs after activation

## Region-Aware Payload Fields

### Currency Fields

All monetary responses include `currency_code`:
- `subtotal_amount` + `currency_code`
- `delivery_fee_amount` + `currency_code`
- `tax_amount` + `currency_code`
- `total_amount` + `currency_code`

### Region Fields

- Restaurant responses include `region_id`, `currency_code`, `timezone`
- Order responses include `region_id`, `currency_code`, `tax_zone_id`
- Menu item prices include `currency_code`

### Timezone Fields

- Restaurant responses include `timezone`
- Scheduled order times validated against restaurant/region timezone

## Tax Breakdown Fields

### Order Response

```json
{
  "subtotal_amount": 1000,
  "delivery_fee_amount": 200,
  "tax_amount": 120,
  "total_amount": 1320,
  "currency_code": "USD",
  "pricing_mode": "tax_exclusive",
  "tax_breakdown": [
    {
      "tax_category": "food",
      "tax_rate_name": "State Sales Tax",
      "tax_rate_bps": 800,
      "taxable_amount": 1000,
      "tax_amount": 80
    },
    {
      "tax_category": "delivery",
      "tax_rate_name": "Delivery Tax",
      "tax_rate_bps": 2000,
      "taxable_amount": 200,
      "tax_amount": 40
    }
  ]
}
```

### Order Item Response

```json
{
  "item_name": "Burger",
  "quantity": 2,
  "unit_price_amount": 500,
  "line_subtotal_amount": 1000,
  "line_tax_amount": 80,
  "line_total_amount": 1080,
  "currency_code": "USD"
}
```

## CRUD Behavior by Role

### Customer CRUD

| Resource | Create | Read | Update | Delete | Scope |
|----------|--------|------|--------|--------|-------|
| users/me | No | Yes (own) | Yes (own) | No | Own profile only |
| user_addresses | Yes | Yes (own) | Yes (own) | Yes (own) | Own addresses only |
| restaurants | No | Yes (all) | No | No | Read-only |
| menus | No | Yes (all) | No | No | Read-only |
| orders | Yes | Yes (own) | No (cancel only) | No | Own orders only |
| payments | No | Yes (own) | No | No | Own payments only |

### Restaurant Owner CRUD

| Resource | Create | Read | Update | Delete | Scope |
|----------|--------|------|--------|--------|-------|
| restaurants | Yes | Yes (own) | Yes (own) | Yes (soft, own) | Own restaurants only |
| restaurant_hours | Yes | Yes (own) | Yes (own) | Yes (own) | Own restaurants only |
| menu_items | Yes | Yes (own) | Yes (own) | Yes (soft, own) | Own restaurants only |
| orders | No | Yes (own) | Yes (workflow only) | No | Own restaurant orders only |

### Courier CRUD

| Resource | Create | Read | Update | Delete | Scope |
|----------|--------|------|--------|--------|-------|
| courier/profile | No | Yes (own) | Yes (limited) | No | Own profile only |
| deliveries | No | Yes (available + own) | Yes (pickup/complete) | No | Available + own only |

### Admin CRUD

| Resource | Create | Read | Update | Delete | Scope |
|----------|--------|------|--------|--------|-------|
| users | Yes | Yes (all) | Yes (limited) | Yes (soft) | All users |
| restaurants | No | Yes (all) | Yes | Yes (soft) | All restaurants |
| regions | Yes | Yes (all) | Yes | Yes (soft) | All regions |
| region_configs | Yes | Yes (all) | Yes | Yes | All configs |
| tax_categories | Yes | Yes (all) | Yes | Yes | All categories |
| tax_zones | Yes | Yes (all) | Yes | Yes | All zones |
| tax_rates | Yes | Yes (all) | Yes | Yes (soft) | All rates |
| audit_logs | No | Yes (all) | No | No | Read-only |
| orders | No | Yes (all) | Yes (limited) | No (cancel only) | All orders |
| payments | No | Yes (all) | Yes (status only) | No | All payments |

## Sample Request/Response Payloads

### Register

**Request**:
```json
POST /v1/auth/register
{
  "email": "customer@example.com",
  "password": "SecurePass123!",
  "role": "customer",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "customer@example.com",
      "role": "customer",
      "first_name": "John",
      "last_name": "Doe"
    },
    "access_token": "jwt_token",
    "refresh_token": "refresh_token"
  }
}
```

### Login

**Request**:
```json
POST /v1/auth/login
{
  "email": "customer@example.com",
  "password": "SecurePass123!"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "access_token": "jwt_token",
    "refresh_token": "refresh_token"
  }
}
```

### Create Order (ASAP)

**Request**:
```json
POST /v1/orders
{
  "restaurant_id": "restaurant_uuid",
  "delivery_type": "asap",
  "delivery_address": {
    "street_address": "123 Main St",
    "city": "San Francisco",
    "state_or_province": "CA",
    "postal_code": "94102",
    "country_code": "US",
    "latitude": 37.7749,
    "longitude": -122.4194
  },
  "items": [
    {
      "menu_item_id": "menu_item_uuid",
      "quantity": 2
    }
  ],
  "notes": "Extra napkins please"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "order_uuid",
    "status": "pending_payment",
    "delivery_type": "asap",
    "currency_code": "USD",
    "subtotal_amount": 1000,
    "delivery_fee_amount": 200,
    "tax_amount": 120,
    "total_amount": 1320,
    "pricing_mode": "tax_exclusive",
    "items": [
      {
        "id": "order_item_uuid",
        "item_name": "Burger",
        "quantity": 2,
        "unit_price_amount": 500,
        "line_subtotal_amount": 1000,
        "line_tax_amount": 80,
        "line_total_amount": 1080,
        "currency_code": "USD"
      }
    ],
    "delivery_address": {
      "street_address": "123 Main St",
      "city": "San Francisco",
      "state_or_province": "CA",
      "postal_code": "94102",
      "country_code": "US"
    },
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

### Create Order (Scheduled)

**Request**:
```json
POST /v1/orders
{
  "restaurant_id": "restaurant_uuid",
  "delivery_type": "scheduled",
  "scheduled_for": "2024-01-01T18:00:00Z",
  "delivery_address": {
    "street_address": "123 Main St",
    "city": "San Francisco",
    "state_or_province": "CA",
    "postal_code": "94102",
    "country_code": "US"
  },
  "items": [
    {
      "menu_item_id": "menu_item_uuid",
      "quantity": 2
    }
  ]
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "order_uuid",
    "status": "pending_payment",
    "delivery_type": "scheduled",
    "scheduled_for": "2024-01-01T18:00:00Z",
    "delivery_window_start": "2024-01-01T17:45:00Z",
    "delivery_window_end": "2024-01-01T18:15:00Z",
    "currency_code": "USD",
    "subtotal_amount": 1000,
    "delivery_fee_amount": 200,
    "tax_amount": 120,
    "total_amount": 1320,
    "items": [...]
  }
}
```

### Create Restaurant

**Request**:
```json
POST /v1/owner/restaurants
{
  "name": "Tasty Burgers",
  "description": "Best burgers in town",
  "cuisine_type": "American",
  "street_address": "456 Food St",
  "city": "San Francisco",
  "state_or_province": "CA",
  "postal_code": "94103",
  "country_code": "US",
  "latitude": 37.7849,
  "longitude": -122.4094,
  "phone": "+14155551234",
  "email": "info@tastyburgers.com"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "restaurant_uuid",
    "name": "Tasty Burgers",
    "description": "Best burgers in town",
    "cuisine_type": "American",
    "currency_code": "USD",
    "timezone": "America/Los_Angeles",
    "is_active": false,
    "is_approved": false,
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

### Create Menu Item

**Request**:
```json
POST /v1/owner/restaurants/{id}/menu-items
{
  "name": "Classic Burger",
  "description": "Beef patty with lettuce, tomato, onion",
  "price_amount": 1200,
  "tax_category_id": "tax_category_uuid"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "menu_item_uuid",
    "name": "Classic Burger",
    "description": "Beef patty with lettuce, tomato, onion",
    "price_amount": 1200,
    "currency_code": "USD",
    "is_available": true,
    "is_active": true
  }
}
```

### Accept Order (Restaurant Owner)

**Request**:
```json
PATCH /v1/owner/orders/{id}/accept
{}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "order_uuid",
    "status": "accepted",
    "updated_at": "2024-01-01T12:05:00Z"
  }
}
```

### Claim Delivery (Courier)

**Request**:
```json
POST /v1/courier/deliveries/{id}/claim
{}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "delivery_uuid",
    "order_id": "order_uuid",
    "courier_id": "courier_uuid",
    "status": "assigned",
    "assigned_at": "2024-01-01T12:10:00Z"
  }
}
```

### Create Tax Rate (Admin)

**Request**:
```json
POST /v1/admin/tax-rates
{
  "tax_zone_id": "tax_zone_uuid",
  "tax_category_id": "tax_category_uuid",
  "name": "California Sales Tax",
  "rate_bps": 825,
  "is_inclusive": false,
  "applies_to_delivery_fee": true,
  "effective_from": "2024-01-01T00:00:00Z"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "tax_rate_uuid",
    "tax_zone_id": "tax_zone_uuid",
    "tax_category_id": "tax_category_uuid",
    "name": "California Sales Tax",
    "rate_bps": 825,
    "is_inclusive": false,
    "applies_to_delivery_fee": true,
    "effective_from": "2024-01-01T00:00:00Z",
    "is_active": true
  }
}
```

## Error Response Contract

### Success Response

```json
{
  "success": true,
  "data": { ... }
}
```

### Error Response

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "invalid request",
    "details": {
      "field": "email",
      "reason": "invalid email format"
    }
  },
  "request_id": "uuid"
}
```

### Error Codes

| Code | Description | HTTP Status |
|------|-------------|-------------|
| VALIDATION_ERROR | Request validation failed | 400 |
| UNAUTHORIZED | No authentication or invalid token | 401 |
| FORBIDDEN | Insufficient permissions | 403 |
| NOT_FOUND | Resource not found | 404 |
| CONFLICT | Idempotency conflict or state conflict | 409 |
| UNPROCESSABLE_ENTITY | Business rule violation | 422 |
| RATE_LIMIT_EXCEEDED | Too many requests | 429 |
| INTERNAL_ERROR | Server error | 500 |

### Common Error Scenarios

**Invalid State Transition**:
```json
{
  "success": false,
  "error": {
    "code": "UNPROCESSABLE_ENTITY",
    "message": "invalid order status transition",
    "details": {
      "current_status": "pending",
      "requested_status": "ready_for_pickup",
      "reason": "order must be accepted before marking ready"
    }
  },
  "request_id": "uuid"
}
```

**Ownership Violation**:
```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "access denied",
    "details": {
      "reason": "you do not have permission to access this resource"
    }
  },
  "request_id": "uuid"
}
```

**Idempotency Conflict**:
```json
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "idempotency key already used",
    "details": {
      "key": "idempotency_key",
      "original_response": { ... }
    }
  },
  "request_id": "uuid"
}
```

## Pagination

List endpoints support pagination via query parameters:

- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20, max: 100)
- `sort`: Sort field (e.g., `created_at`, `name`)
- `order`: Sort order (asc, desc, default: desc)

**Example**:
```
GET /v1/me/orders?page=1&limit=20&sort=created_at&order=desc
```

**Response**:
```json
{
  "success": true,
  "data": {
    "items": [...],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

## Rate Limiting

Rate limiting is enforced on:

- Authentication endpoints (register, login, refresh)
- Order creation
- Payment initiation
- Courier claim

Rate limits by IP address and optionally by user.

Headers:
- `X-RateLimit-Limit`: Request limit per window
- `X-RateLimit-Remaining`: Remaining requests in window
- `X-RateLimit-Reset`: Unix timestamp when limit resets

## Idempotency

Critical POST endpoints support idempotency via `Idempotency-Key` header:

- Order creation
- Payment initiation

**Request**:
```
POST /v1/orders
Idempotency-Key: unique-key-per-request
```

**Response**:
```
X-Idempotency-Key: unique-key-per-request
```

Duplicate requests with same key return original response.

## Summary

The API is designed with:
- Clear role-based access control
- Strict workflow state enforcement
- Region-aware behavior
- Tax-aware pricing with breakdowns
- Comprehensive error handling
- Pagination support
- Rate limiting
- Idempotency for critical operations
- Consistent response contracts
