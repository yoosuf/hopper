# Database Design

## Overview

This document describes the database schema, indexes, constraints, and data model for the Food Delivery API Backend. The database is PostgreSQL 16+ with UUID primary keys, timestamps, and strong referential integrity.

## ERD Overview

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   users     │────▶│refresh_tokens│     │   regions   │
└─────────────┘     └──────────────┘     └─────────────┘
       │                                         │
       │                                         │
       ▼                                         ▼
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│user_addresses│    │ restaurants  │────▶│region_configs│
└─────────────┘     └──────────────┘     └─────────────┘
                          │
                          │
                          ▼
                    ┌──────────────┐
                    │restaurant_hours│
                    └──────────────┘
                          │
                          │
                          ▼
                    ┌──────────────┐
                    │  menu_items  │
                    └──────────────┘
                          │
                          │
                          ▼
                    ┌──────────────┐     ┌─────────────┐
                    │    orders    │────▶│  deliveries │
                    └──────────────┘     └─────────────┘
                          │
                          │
                          ▼
                    ┌──────────────┐
                    │ order_items  │
                    └──────────────┘
                          │
                          │
                          ▼
                    ┌──────────────┐
                    │ order_tax_lines│
                    └──────────────┘

┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  payments   │     │ outbox_events│     │   jobs      │
└─────────────┘     └──────────────┘     └─────────────┘

┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│ audit_logs  │     │idempotency_keys│   │ tax_zones   │
└─────────────┘     └──────────────┘     └─────────────┘
                                                │
                                                │
                                                ▼
                                          ┌─────────────┐
                                          │ tax_rates   │
                                          └─────────────┘
                                                │
                                                │
                                                ▼
                                          ┌─────────────┐
                                          │tax_categories│
                                          └─────────────┘
```

## Table-by-Table Breakdown

### users

User accounts and authentication data.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| email | VARCHAR(255) UNIQUE NOT NULL | User email address |
| password_hash | VARCHAR(255) NOT NULL | Bcrypt/argon2 hashed password |
| role | VARCHAR(50) NOT NULL | Role: customer, restaurant_owner, courier, admin |
| first_name | VARCHAR(100) | First name |
| last_name | VARCHAR(100) | Last name |
| phone | VARCHAR(20) | Phone number |
| is_active | BOOLEAN NOT NULL DEFAULT true | Account active flag |
| email_verified | BOOLEAN NOT NULL DEFAULT false | Email verification status |
| deleted_at | TIMESTAMP WITH TIME ZONE | Soft delete timestamp |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Update timestamp |

**Indexes**:
- `idx_users_email` on (email)
- `idx_users_role` on (role)
- `idx_users_is_active` on (is_active) WHERE deleted_at IS NULL

**Constraints**:
- CHECK (role IN ('customer', 'restaurant_owner', 'courier', 'admin'))
- CHECK (is_active = true OR deleted_at IS NOT NULL)

### refresh_tokens

JWT refresh tokens for authentication.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| user_id | UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE | User ID |
| token_hash | VARCHAR(255) UNIQUE NOT NULL | Hashed refresh token |
| expires_at | TIMESTAMP WITH TIME ZONE NOT NULL | Expiration timestamp |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |
| revoked_at | TIMESTAMP WITH TIME ZONE | Revocation timestamp |

**Indexes**:
- `idx_refresh_tokens_user_id` on (user_id)
- `idx_refresh_tokens_token_hash` on (token_hash)
- `idx_refresh_tokens_expires_at` on (expires_at) WHERE revoked_at IS NULL

### user_addresses

Customer delivery addresses.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| user_id | UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE | User ID |
| label | VARCHAR(50) | Address label (home, work, etc.) |
| street_address | VARCHAR(255) NOT NULL | Street address |
| city | VARCHAR(100) NOT NULL | City |
| state_or_province | VARCHAR(100) | State or province |
| postal_code | VARCHAR(20) NOT NULL | Postal code |
| country_code | VARCHAR(2) NOT NULL | ISO country code |
| latitude | DECIMAL(10, 8) | Latitude |
| longitude | DECIMAL(11, 8) | Longitude |
| is_default | BOOLEAN NOT NULL DEFAULT false | Default address flag |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Update timestamp |

**Indexes**:
- `idx_user_addresses_user_id` on (user_id)
- `idx_user_addresses_is_default` on (user_id, is_default) WHERE is_default = true

### regions

Geographic regions for multi-region support.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| code | VARCHAR(10) UNIQUE NOT NULL | Region code (e.g., US-NY, EU-GB) |
| name | VARCHAR(100) NOT NULL | Region name |
| country_code | VARCHAR(2) NOT NULL | ISO country code |
| timezone | VARCHAR(50) NOT NULL | IANA timezone (e.g., America/New_York) |
| currency_code | VARCHAR(3) NOT NULL | ISO currency code (e.g., USD, GBP) |
| is_active | BOOLEAN NOT NULL DEFAULT true | Active flag |
| deleted_at | TIMESTAMP WITH TIME ZONE | Soft delete timestamp |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Update timestamp |

**Indexes**:
- `idx_regions_code` on (code)
- `idx_regions_country_code` on (country_code)
- `idx_regions_is_active` on (is_active) WHERE deleted_at IS NULL

### region_configs

Platform configuration per region.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| region_id | UUID NOT NULL REFERENCES regions(id) ON DELETE CASCADE | Region ID |
| platform_fee_basis_points | INTEGER NOT NULL DEFAULT 0 | Platform fee in basis points (100 = 1%) |
| default_delivery_window_minutes | INTEGER NOT NULL DEFAULT 30 | Default delivery window in minutes |
| order_activation_lead_minutes | INTEGER NOT NULL DEFAULT 15 | Lead time for order activation |
| allow_scheduled_orders | BOOLEAN NOT NULL DEFAULT true | Allow scheduled orders |
| delivery_fee_taxable_default | BOOLEAN NOT NULL DEFAULT true | Delivery fee taxable by default |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Update timestamp |

**Indexes**:
- `idx_region_configs_region_id` on (region_id) UNIQUE

### tax_zones

Tax jurisdictions within regions.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| region_id | UUID NOT NULL REFERENCES regions(id) ON DELETE CASCADE | Region ID |
| code | VARCHAR(50) NOT NULL | Tax zone code |
| name | VARCHAR(100) NOT NULL | Tax zone name |
| country_code | VARCHAR(2) NOT NULL | ISO country code |
| state_or_province | VARCHAR(100) | State or province |
| city | VARCHAR(100) | City |
| postal_code_pattern | VARCHAR(100) | Postal code regex pattern |
| is_active | BOOLEAN NOT NULL DEFAULT true | Active flag |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Update timestamp |

**Indexes**:
- `idx_tax_zones_region_id` on (region_id)
- `idx_tax_zones_code` on (code)
- `idx_tax_zones_is_active` on (is_active)

### tax_categories

Product categories for tax classification.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| code | VARCHAR(50) UNIQUE NOT NULL | Tax category code |
| name | VARCHAR(100) NOT NULL | Tax category name |
| description | TEXT | Description |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Update timestamp |

**Indexes**:
- `idx_tax_categories_code` on (code)

### tax_rates

Tax rates per tax zone and category.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| tax_zone_id | UUID NOT NULL REFERENCES tax_zones(id) ON DELETE CASCADE | Tax zone ID |
| tax_category_id | UUID NOT NULL REFERENCES tax_categories(id) ON DELETE CASCADE | Tax category ID |
| name | VARCHAR(100) NOT NULL | Tax rate name |
| rate_bps | INTEGER NOT NULL | Rate in basis points (1000 = 10%) |
| is_inclusive | BOOLEAN NOT NULL DEFAULT false | Tax-inclusive pricing |
| applies_to_delivery_fee | BOOLEAN NOT NULL DEFAULT false | Applies to delivery fee |
| effective_from | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Effective from date |
| effective_to | TIMESTAMP WITH TIME ZONE | Effective to date |
| priority | INTEGER NOT NULL DEFAULT 0 | Priority for overlapping rates |
| is_active | BOOLEAN NOT NULL DEFAULT true | Active flag |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Update timestamp |

**Indexes**:
- `idx_tax_rates_tax_zone_id` on (tax_zone_id)
- `idx_tax_rates_tax_category_id` on (tax_category_id)
- `idx_tax_rates_is_active` on (is_active)
- `idx_tax_rates_effective_dates` on (effective_from, effective_to)

**Constraints**:
- CHECK (rate_bps >= 0 AND rate_bps <= 10000)

### restaurants

Restaurant profiles and business information.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| owner_id | UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT | Owner user ID |
| region_id | UUID NOT NULL REFERENCES regions(id) ON DELETE RESTRICT | Region ID |
| name | VARCHAR(255) NOT NULL | Restaurant name |
| description | TEXT | Description |
| cuisine_type | VARCHAR(100) | Cuisine type |
| street_address | VARCHAR(255) NOT NULL | Street address |
| city | VARCHAR(100) NOT NULL | City |
| state_or_province | VARCHAR(100) | State or province |
| postal_code | VARCHAR(20) NOT NULL | Postal code |
| country_code | VARCHAR(2) NOT NULL | ISO country code |
| latitude | DECIMAL(10, 8) | Latitude |
| longitude | DECIMAL(11, 8) | Longitude |
| phone | VARCHAR(20) | Phone number |
| email | VARCHAR(255) | Email |
| currency_code | VARCHAR(3) NOT NULL | ISO currency code |
| timezone | VARCHAR(50) NOT NULL | IANA timezone |
| is_active | BOOLEAN NOT NULL DEFAULT false | Active flag (requires approval) |
| is_approved | BOOLEAN NOT NULL DEFAULT false | Admin approval flag |
| deleted_at | TIMESTAMP WITH TIME ZONE | Soft delete timestamp |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Update timestamp |

**Indexes**:
- `idx_restaurants_owner_id` on (owner_id)
- `idx_restaurants_region_id` on (region_id)
- `idx_restaurants_is_active` on (is_active) WHERE deleted_at IS NULL
- `idx_restaurants_is_approved` on (is_approved) WHERE deleted_at IS NULL
- `idx_restaurants_location` on (latitude, longitude) WHERE latitude IS NOT NULL

### restaurant_hours

Restaurant operating hours.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| restaurant_id | UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE | Restaurant ID |
| day_of_week | INTEGER NOT NULL | Day of week (0 = Sunday, 6 = Saturday) |
| open_time | TIME NOT NULL | Opening time |
| close_time | TIME NOT NULL | Closing time |
| is_closed | BOOLEAN NOT NULL DEFAULT false | Closed on this day |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Update timestamp |

**Indexes**:
- `idx_restaurant_hours_restaurant_id` on (restaurant_id)
- `idx_restaurant_hours_day_of_week` on (restaurant_id, day_of_week)

**Constraints**:
- CHECK (day_of_week >= 0 AND day_of_week <= 6)
- CHECK (open_time < close_time OR is_closed = true)
- UNIQUE (restaurant_id, day_of_week)

### menu_items

Menu items for restaurants.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| restaurant_id | UUID NOT NULL REFERENCES restaurants(id) ON DELETE RESTRICT | Restaurant ID |
| tax_category_id | UUID REFERENCES tax_categories(id) ON DELETE SET NULL | Tax category ID |
| name | VARCHAR(255) NOT NULL | Item name |
| description | TEXT | Description |
| price_amount | INTEGER NOT NULL | Price in minor units |
| currency_code | VARCHAR(3) NOT NULL | ISO currency code |
| is_available | BOOLEAN NOT NULL DEFAULT true | Available flag |
| is_active | BOOLEAN NOT NULL DEFAULT true | Active flag |
| deleted_at | TIMESTAMP WITH TIME ZONE | Soft delete timestamp |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Update timestamp |

**Indexes**:
- `idx_menu_items_restaurant_id` on (restaurant_id)
- `idx_menu_items_tax_category_id` on (tax_category_id)
- `idx_menu_items_is_active` on (is_active) WHERE deleted_at IS NULL
- `idx_menu_items_is_available` on (restaurant_id, is_available) WHERE is_active = true AND deleted_at IS NULL

**Constraints**:
- CHECK (price_amount >= 0)

### orders

Customer orders with pricing and tax snapshots.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| customer_id | UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT | Customer user ID |
| restaurant_id | UUID NOT NULL REFERENCES restaurants(id) ON DELETE RESTRICT | Restaurant ID |
| region_id | UUID NOT NULL REFERENCES regions(id) ON DELETE RESTRICT | Region ID |
| tax_zone_id | UUID REFERENCES tax_zones(id) ON DELETE SET NULL | Tax zone ID |
| currency_code | VARCHAR(3) NOT NULL | ISO currency code |
| status | VARCHAR(50) NOT NULL | Order status |
| delivery_type | VARCHAR(20) NOT NULL | Delivery type: asap, scheduled |
| scheduled_for | TIMESTAMP WITH TIME ZONE | Scheduled delivery time |
| activated_at | TIMESTAMP WITH TIME ZONE | Activation timestamp |
| delivery_window_start | TIMESTAMP WITH TIME ZONE | Delivery window start |
| delivery_window_end | TIMESTAMP WITH TIME ZONE | Delivery window end |
| delivery_address_street | VARCHAR(255) NOT NULL | Delivery address (snapshot) |
| delivery_address_city | VARCHAR(100) NOT NULL | Delivery city (snapshot) |
| delivery_address_state | VARCHAR(100) | Delivery state (snapshot) |
| delivery_address_postal_code | VARCHAR(20) NOT NULL | Delivery postal code (snapshot) |
| delivery_address_country_code | VARCHAR(2) NOT NULL | Delivery country (snapshot) |
| delivery_address_latitude | DECIMAL(10, 8) | Delivery latitude (snapshot) |
| delivery_address_longitude | DECIMAL(11, 8) | Delivery longitude (snapshot) |
| subtotal_amount | INTEGER NOT NULL | Subtotal in minor units |
| delivery_fee_amount | INTEGER NOT NULL | Delivery fee in minor units |
| tax_amount | INTEGER NOT NULL | Tax amount in minor units |
| total_amount | INTEGER NOT NULL | Total amount in minor units |
| pricing_mode | VARCHAR(20) NOT NULL | Pricing mode: tax_inclusive, tax_exclusive |
| notes | TEXT | Customer notes |
| cancellation_reason | TEXT | Cancellation reason |
| cancelled_at | TIMESTAMP WITH TIME ZONE | Cancellation timestamp |
| cancelled_by_id | UUID REFERENCES users(id) | Cancelled by user ID |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Update timestamp |

**Indexes**:
- `idx_orders_customer_id` on (customer_id)
- `idx_orders_restaurant_id` on (restaurant_id)
- `idx_orders_region_id` on (region_id)
- `idx_orders_status` on (status)
- `idx_orders_scheduled_for` on (scheduled_for) WHERE status = 'scheduled'
- `idx_orders_created_at` on (created_at)

**Constraints**:
- CHECK (status IN ('pending_payment', 'scheduled', 'pending', 'accepted', 'rejected', 'preparing', 'ready_for_pickup', 'picked_up', 'delivered', 'cancelled'))
- CHECK (delivery_type IN ('asap', 'scheduled'))
- CHECK (pricing_mode IN ('tax_inclusive', 'tax_exclusive'))
- CHECK (subtotal_amount >= 0)
- CHECK (delivery_fee_amount >= 0)
- CHECK (tax_amount >= 0)
- CHECK (total_amount >= 0)

### order_items

Line items for orders with price snapshots.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| order_id | UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE | Order ID |
| menu_item_id | UUID REFERENCES menu_items(id) ON DELETE SET NULL | Menu item ID |
| item_name | VARCHAR(255) NOT NULL | Item name (snapshot) |
| quantity | INTEGER NOT NULL | Quantity |
| unit_price_amount | INTEGER NOT NULL | Unit price in minor units (snapshot) |
| line_subtotal_amount | INTEGER NOT NULL | Line subtotal in minor units |
| line_tax_amount | INTEGER NOT NULL | Line tax amount in minor units |
| line_total_amount | INTEGER NOT NULL | Line total in minor units |
| currency_code | VARCHAR(3) NOT NULL | ISO currency code (snapshot) |
| tax_category_id | UUID REFERENCES tax_categories(id) | Tax category ID (snapshot) |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |

**Indexes**:
- `idx_order_items_order_id` on (order_id)
- `idx_order_items_menu_item_id` on (menu_item_id)

**Constraints**:
- CHECK (quantity > 0)
- CHECK (unit_price_amount >= 0)
- CHECK (line_subtotal_amount >= 0)
- CHECK (line_tax_amount >= 0)
- CHECK (line_total_amount >= 0)

### order_tax_lines

Tax breakdown for orders.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| order_id | UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE | Order ID |
| tax_category_id | UUID REFERENCES tax_categories(id) | Tax category ID |
| tax_rate_name | VARCHAR(100) NOT NULL | Tax rate name (snapshot) |
| tax_rate_bps | INTEGER NOT NULL | Tax rate in basis points (snapshot) |
| is_inclusive | BOOLEAN NOT NULL | Tax-inclusive flag (snapshot) |
| taxable_amount | INTEGER NOT NULL | Taxable amount in minor units |
| tax_amount | INTEGER NOT NULL | Tax amount in minor units |
| currency_code | VARCHAR(3) NOT NULL | ISO currency code (snapshot) |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |

**Indexes**:
- `idx_order_tax_lines_order_id` on (order_id)
- `idx_order_tax_lines_tax_category_id` on (tax_category_id)

### deliveries

Delivery assignments and status.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| order_id | UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE | Order ID |
| courier_id | UUID REFERENCES users(id) ON DELETE SET NULL | Courier user ID |
| status | VARCHAR(50) NOT NULL | Delivery status |
| assigned_at | TIMESTAMP WITH TIME ZONE | Assignment timestamp |
| picked_up_at | TIMESTAMP WITH TIME ZONE | Pickup timestamp |
| delivered_at | TIMESTAMP WITH TIME ZONE | Delivery timestamp |
| failed_at | TIMESTAMP WITH TIME ZONE | Failure timestamp |
| failure_reason | TEXT | Failure reason |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Update timestamp |

**Indexes**:
- `idx_deliveries_order_id` on (order_id) UNIQUE
- `idx_deliveries_courier_id` on (courier_id)
- `idx_deliveries_status` on (status)

**Constraints**:
- CHECK (status IN ('pending_schedule', 'unassigned', 'assigned', 'picked_up', 'delivered', 'failed', 'cancelled'))

### payments

Payment records with provider information.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| order_id | UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT | Order ID |
| provider | VARCHAR(50) NOT NULL | Payment provider (stripe, mock, etc.) |
| status | VARCHAR(50) NOT NULL | Payment status |
| amount | INTEGER NOT NULL | Amount in minor units |
| currency_code | VARCHAR(3) NOT NULL | ISO currency code |
| provider_reference | VARCHAR(255) | Provider reference ID |
| metadata | JSONB | Provider-specific metadata |
| failure_reason | TEXT | Failure reason |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Update timestamp |

**Indexes**:
- `idx_payments_order_id` on (order_id) UNIQUE
- `idx_payments_provider` on (provider)
- `idx_payments_status` on (status)

**Constraints**:
- CHECK (status IN ('pending', 'processing', 'succeeded', 'failed', 'cancelled'))
- CHECK (amount >= 0)

### outbox_events

Outbox events for reliable async side effects.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| event_type | VARCHAR(100) NOT NULL | Event type |
| aggregate_type | VARCHAR(100) NOT NULL | Aggregate type (order, payment, etc.) |
| aggregate_id | UUID NOT NULL | Aggregate ID |
| payload | JSONB NOT NULL | Event payload |
| processed | BOOLEAN NOT NULL DEFAULT false | Processed flag |
| processed_at | TIMESTAMP WITH TIME ZONE | Processing timestamp |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |

**Indexes**:
- `idx_outbox_events_processed` on (processed, created_at) WHERE processed = false
- `idx_outbox_events_aggregate` on (aggregate_type, aggregate_id)

### jobs

Background job queue for workers.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| job_type | VARCHAR(100) NOT NULL | Job type |
| payload | JSONB NOT NULL | Job payload |
| status | VARCHAR(50) NOT NULL | Job status |
| retry_count | INTEGER NOT NULL DEFAULT 0 | Retry count |
| max_retries | INTEGER NOT NULL DEFAULT 3 | Max retries |
| next_attempt_at | TIMESTAMP WITH TIME ZONE NOT NULL | Next attempt timestamp |
| locked_at | TIMESTAMP WITH TIME ZONE | Lock timestamp |
| locked_by | VARCHAR(255) | Worker identifier |
| processed_at | TIMESTAMP WITH TIME ZONE | Processing timestamp |
| failed_at | TIMESTAMP WITH TIME ZONE | Failure timestamp |
| last_error | TEXT | Last error message |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |

**Indexes**:
- `idx_jobs_status_next_attempt` on (status, next_attempt_at) WHERE status IN ('pending', 'failed')
- `idx_jobs_type` on (job_type)

**Constraints**:
- CHECK (status IN ('pending', 'processing', 'succeeded', 'failed', 'dead_letter'))
- CHECK (retry_count >= 0)
- CHECK (max_retries >= 0)

### audit_logs

Audit log for sensitive actions.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| actor_id | UUID NOT NULL | Actor user ID |
| actor_role | VARCHAR(50) NOT NULL | Actor role |
| action | VARCHAR(100) NOT NULL | Action performed |
| entity_type | VARCHAR(100) NOT NULL | Entity type |
| entity_id | UUID | Entity ID |
| metadata | JSONB | Additional metadata |
| ip_address | INET | Actor IP address |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |

**Indexes**:
- `idx_audit_logs_actor_id` on (actor_id)
- `idx_audit_logs_entity` on (entity_type, entity_id)
- `idx_audit_logs_created_at` on (created_at)

### idempotency_keys

Idempotency key management for critical endpoints.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| key | VARCHAR(255) NOT NULL | Idempotency key |
| scope | VARCHAR(100) NOT NULL | Scope (user_id, endpoint, etc.) |
| request_hash | VARCHAR(64) | Request fingerprint hash |
| response_code | INTEGER | HTTP response code |
| response_body | JSONB | Response body |
| expires_at | TIMESTAMP WITH TIME ZONE NOT NULL | Expiration timestamp |
| created_at | TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() | Creation timestamp |

**Indexes**:
- `idx_idempotency_keys_key_scope` on (key, scope) UNIQUE
- `idx_idempotency_keys_expires_at` on (expires_at) WHERE expires_at > now()

## Index Strategy

### Foreign Key Indexes

All foreign key columns are indexed for join performance:
- user_id, restaurant_id, region_id, order_id, etc.

### Query Pattern Indexes

Indexes are created based on common query patterns:
- User listings by role and active status
- Restaurant listings by region and approval status
- Order listings by customer, restaurant, status
- Delivery listings by courier and status
- Tax rates by zone, category, and active status

### Partial Indexes

Partial indexes are used where appropriate to reduce index size:
- Active records only (WHERE deleted_at IS NULL)
- Available menu items only (WHERE is_available = true)
- Unprocessed outbox events only (WHERE processed = false)
- Due jobs only (WHERE status IN ('pending', 'failed'))

## Constraints

### Check Constraints

Check constraints enforce data integrity:
- Role values limited to valid roles
- Status values limited to valid statuses
- Monetary values non-negative
- Day of week range 0-6
- Time validation (open < close unless closed)

### Foreign Key Constraints

Foreign keys ensure referential integrity:
- ON DELETE CASCADE for dependent records (refresh_tokens, user_addresses)
- ON DELETE RESTRICT for critical business records (orders, payments)
- ON DELETE SET NULL for optional references (tax_category_id, courier_id)

### Unique Constraints

Unique constraints prevent duplicates:
- Email addresses
- Refresh token hashes
- Region codes
- Tax category codes
- Restaurant hours per day
- Idempotency keys per scope

## Transaction Notes

### Critical Write Transactions

The following operations use transactions:

1. **Order Creation**: Single transaction for order, order_items, delivery, payment, outbox_event
2. **Order Status Transitions**: Update order status, delivery status, create outbox event
3. **Payment Status Updates**: Update payment, order status, create audit log
4. **Courier Claim**: Update delivery assignment, create outbox event
5. **Restaurant Approval**: Update restaurant status, create audit log
6. **Tax Configuration Changes**: Update tax records, create audit log

### Isolation Level

Default `READ COMMITTED` isolation is used:
- Prevents dirty reads
- Allows non-repeatable reads (acceptable for this use case)
- Good balance between consistency and performance

### Locking Strategy

- `SELECT FOR UPDATE SKIP LOCKED` for job claiming
- Row-level locks for delivery assignment
- No explicit locks for most operations (rely on transaction isolation)

## Job/Outbox Table Explanation

### outbox_events Table

Purpose: Reliable async side effects without message broker.

Design:
- Created in same transaction as triggering change
- Contains event type, aggregate info, payload
- Workers poll for unprocessed events
- Marked as processed after successful handling
- No deletion (audit trail)

Benefits:
- Exactly-once semantics with transactional outbox
- No message broker infrastructure needed initially
- Audit trail of all events

### jobs Table

Purpose: Background job queue for workers.

Design:
- Polling-based job queue
- Retry count and max retries
- Next attempt timestamp for scheduling
- Lock timestamp and worker ID for concurrency
- Status tracking (pending, processing, succeeded, failed, dead_letter)

Benefits:
- Simple to implement with database
- Safe concurrent processing with SKIP LOCKED
- Retry logic built-in
- Dead-letter queue for failed jobs

## Schedule-Related Columns Explanation

### Orders Table Schedule Columns

- `delivery_type`: 'asap' or 'scheduled'
- `scheduled_for`: UTC timestamp for scheduled delivery
- `activated_at`: When scheduled order became active
- `delivery_window_start`: Expected delivery window start
- `delivery_window_end`: Expected delivery window end

### Scheduled Order Lifecycle

1. Customer creates order with `delivery_type = 'scheduled'` and `scheduled_for` in future
2. Order status = 'pending_payment' → 'scheduled' after payment
3. Worker activates order when `scheduled_for <= now()`
4. Worker sets `activated_at` and status = 'pending'
5. Normal fulfillment flow continues

### Timezone Considerations

- All timestamps stored in UTC
- Region timezone stored in `regions.timezone`
- Restaurant timezone stored in `restaurants.timezone`
- Validation converts UTC to local time for business hour checks
- Worker activation uses UTC comparison but validates local hours

## Region Tables

### regions Table

Core region definition with timezone and currency.

### region_configs Table

Platform settings per region:
- Platform fee basis points
- Default delivery window
- Scheduled order lead time
- Delivery fee taxability

### tax_zones Table

Tax jurisdictions within regions:
- Can be country-wide, state-wide, or city-specific
- Postal code patterns for fine-grained matching
- Links to tax_rates

## Tax Tables

### tax_categories Table

Product categories for tax classification:
- Food, beverages, alcohol, etc.
- Different tax rates per category

### tax_rates Table

Tax rates per zone and category:
- Rate in basis points (1000 = 10%)
- Inclusive vs exclusive tax
- Applies to delivery fee flag
- Effective date ranges
- Priority for overlapping rates

### order_tax_lines Table

Tax breakdown snapshot per order:
- Preserves tax calculation at order time
- Enables invoice generation
- Tax changes don't affect historical orders

## Money Storage Strategy

### Integer Minor Units

All monetary values stored as integers in minor units:
- USD: cents (100 cents = $1.00)
- GBP: pence (100 pence = £1.00)
- EUR: cents (100 cents = €1.00)

### Currency Code

Currency code stored alongside all monetary values:
- `currency_code` on orders, payments, restaurants
- Snapshotted on orders and order_items
- Prevents currency confusion

### Price Snapshots

Prices snapshotted at order creation:
- `unit_price_amount` on order_items
- `line_subtotal_amount`, `line_tax_amount`, `line_total_amount`
- Menu price changes don't affect historical orders

## Tax Snapshot Strategy

### Order-Level Snapshot

- `tax_amount` on orders (total tax)
- `pricing_mode` (inclusive vs exclusive)
- `tax_zone_id` (which tax zone applied)

### Line-Level Snapshot

- `line_tax_amount` on order_items
- `tax_category_id` on order_items
- Preserves per-line tax calculation

### Tax Breakdown Table

- `order_tax_lines` table stores detailed tax breakdown
- Tax rate name, rate, taxable amount, tax amount
- Enables invoice generation
- Tax rate changes don't affect historical orders

## Soft Delete/Archive Strategy

### Soft Delete Pattern

Uses `deleted_at` timestamp for soft delete:
- NULL = active record
- Non-NULL = deleted/archived record
- Queries filter out deleted records by default

### Soft Delete Applied To

- users (disable/suspend instead of hard delete)
- restaurants (archive instead of hard delete)
- menu_items (archive instead of hard delete if referenced)
- regions (deactivate instead of hard delete)
- tax_rates (deactivate instead of hard delete if used)

### Hard Delete Allowed For

- restaurant_hours (if not historically relevant)
- user_addresses (by owner if not referenced by active orders)

### Never Hard Delete

- orders
- payments
- deliveries
- audit_logs
- outbox_events (may cleanup old records but not hard delete)

## Summary

The database design prioritizes:
- Data integrity with foreign keys and constraints
- Performance with appropriate indexes
- Auditability with immutable historical records
- Multi-region support with region-aware tables
- Tax compliance with flexible tax configuration
- Monetary precision with integer minor units
- Soft delete for business entities
- Hard delete prevention for historical records
