# Architecture

## System Overview

The Food Delivery API Backend is a modular monolith built with Go 1.25+ and PostgreSQL 16+. It is designed as a production baseline with strong security, strict workflow enforcement, multi-region support, tax-aware pricing, and comprehensive auditability.

The system consists of:

- **HTTP API Server** (`cmd/api/main.go`): Handles REST API requests
- **Background Worker** (`cmd/worker/main.go`): Processes async jobs and scheduled tasks
- **PostgreSQL Database**: Stores all persistent data
- **Domain Modules**: Self-contained business logic for each domain
- **Platform Layer**: Shared infrastructure (config, DB, logging, middleware, etc.)

## Module Boundaries

The system is organized into clear domain modules with explicit boundaries:

### Core Domain Modules

- **auth**: Authentication (JWT tokens, refresh tokens, password hashing)
- **users**: User profiles, addresses, role management
- **regions**: Geographic regions, timezones, currencies, regional configuration
- **tax**: Tax zones, tax rates, tax categories, tax calculation
- **restaurants**: Restaurant management, hours, approval workflow
- **menus**: Menu items, categories, availability
- **orders**: Order creation, workflow state machine, cancellation
- **delivery**: Delivery assignment, courier claiming, pickup/complete
- **payments**: Payment processing, provider abstraction, status tracking
- **notifications**: Notification sending (email, push, SMS abstraction)
- **admin**: Administrative operations, overrides, approvals
- **audit**: Audit logging for sensitive actions
- **jobs**: Job queue management, outbox event processing
- **worker**: Background job processing, scheduled task execution

### Platform Modules

- **config**: Configuration loading from environment variables
- **db**: Database connection pooling, transaction management
- **logger**: Structured logging with redaction
- **httpx**: HTTP response helpers, error formatting
- **middleware**: Auth, RBAC, rate limiting, request ID, recovery
- **validator**: Request validation
- **idempotency**: Idempotency key management
- **clock**: Time abstraction for testing
- **metrics**: Metrics collection (Prometheus-style hooks)

## Request Flow

### Typical API Request Flow

```
HTTP Request
  ↓
Middleware Chain (request ID, auth, RBAC, rate limit, idempotency)
  ↓
Handler (validation, DTO mapping)
  ↓
Service (business logic, workflow enforcement)
  ↓
Repository (SQL queries, data access)
  ↓
Database
  ↓
Repository (results)
  ↓
Service (domain transformations, outbox event creation)
  ↓
Transaction Commit
  ↓
Handler (response DTO mapping)
  ↓
HTTP Response
```

### Order Creation Flow (Detailed)

```
POST /v1/orders
  ↓
[Middleware] Auth check (customer role)
[Middleware] RBAC check (customer can create orders)
[Middleware] Rate limit check
[Middleware] Idempotency check (if idempotency key provided)
  ↓
Handler validates request DTO
  ↓
Service validates:
  - Restaurant exists and is active
  - Restaurant belongs to a region
  - All menu items exist and are available
  - Delivery address resolves to a region
  - Scheduled time is valid (if scheduled order)
  - Restaurant is open (if ASAP order)
  ↓
Service calculates:
  - Subtotal from menu item prices (server-side)
  - Delivery fee (region-based)
  - Tax via tax service (region/tax zone based)
  - Total amount
  ↓
Service begins transaction
  ↓
Repository creates:
  - Order record (snapshots: currency, region, tax zone, pricing)
  - Order item records (snapshots: item name, unit price, line totals)
  - Delivery record (initial status)
  - Payment record (pending status)
  - Outbox event (order created)
  ↓
Service commits transaction
  ↓
Handler returns order response DTO
  ↓
[Worker] Will process outbox event asynchronously
```

## Transaction Boundaries

### Critical Write Transactions

The following operations use database transactions to ensure atomicity:

1. **Order Creation**: Creates order, order_items, delivery, payment, outbox event
2. **Order Status Transitions**: Updates order status, delivery status, creates outbox event
3. **Payment Status Updates**: Updates payment status, order status, creates audit log
4. **Courier Claim**: Updates delivery assignment, creates outbox event, prevents double-claim
5. **Restaurant Approval**: Updates restaurant status, creates audit log
6. **Admin Overrides**: Updates entity status, creates audit log
7. **Tax Configuration Changes**: Updates tax records, creates audit log
8. **Region Configuration Changes**: Updates region config, creates audit log

### Transaction Isolation

- Use `READ COMMITTED` isolation level (PostgreSQL default)
- Use `SELECT FOR UPDATE SKIP LOCKED` for concurrent job claiming
- Use row-level locking for delivery assignment to prevent double-claim
- Transactions are kept short to minimize lock contention

### Outbox Pattern

- Outbox events created within the same transaction as the triggering change
- Outbox events include: event type, entity ID, payload, created_at
- Workers poll for unprocessed outbox events
- Workers process events only after transaction commit
- Workers mark events as processed after successful handling
- Failed events retry with exponential backoff
- Max retries before dead-letter marking

## Background Worker Flow

### Worker Architecture

```
Worker Process
  ↓
Polling Loop (every 5-10 seconds)
  ↓
Fetch due jobs (SKIP LOCKED for concurrency)
  ↓
Process each job:
  - Validate job state
  - Execute job handler
  - Mark success or failure
  - Schedule retry if failed
  - Mark dead-letter if max retries exceeded
  ↓
Sleep until next poll
```

### Worker Types

1. **Outbox Event Processor**
   - Polls for unprocessed outbox events
   - Dispatches to appropriate event handlers
   - Handles: order_created, order_status_changed, payment_completed, delivery_assigned, etc.
   - Sends notifications, updates external systems

2. **Scheduled Order Activator**
   - Polls for scheduled orders due for activation
   - Activates orders when `scheduled_for` <= current time
   - Validates restaurant is still open
   - Transitions order from `scheduled` to `pending`
   - Creates outbox event for activation

3. **Delayed Notification Sender**
   - Polls for delayed notification jobs
   - Sends notifications at scheduled times
   - Handles: order reminders, delivery ETA updates

4. **Payment Reconciler**
   - Polls for pending payments
   - Checks payment status with mock provider
   - Updates payment status based on provider response
   - Handles payment failures and retries

5. **Maintenance Jobs**
   - Cleans up old processed outbox events
   - Cleans up old audit logs (if retention policy)
   - Refreshes materialized views (if any)

### Worker Safety

- Workers only process committed transactions
- Workers validate job state before processing
- Workers handle duplicate job processing gracefully
- Workers support graceful shutdown (SIGTERM handling)
- Workers log all processing attempts
- Multiple workers can run safely (SKIP LOCKED)

## Scheduled Order Lifecycle

### Scheduled Order States

```
pending_payment → scheduled → pending → accepted → preparing → ready_for_pickup → picked_up → delivered
```

### Scheduled Order Flow

1. **Customer creates scheduled order**
   - Sets `delivery_type = "scheduled"`
   - Sets `scheduled_for` to future timestamp (in restaurant timezone)
   - Sets `delivery_window_start` and `delivery_window_end`
   - Order status = `pending_payment`

2. **Payment completes**
   - Order status transitions to `scheduled`
   - Order remains inactive until activation time
   - Restaurant and courier cannot see/act on order

3. **Worker activates scheduled order**
   - Worker polls for scheduled orders where `scheduled_for <= now`
   - Validates restaurant is still open
   - Transitions order status to `pending`
   - Sets `activated_at` timestamp
   - Creates outbox event for activation
   - Restaurant can now see and accept order

4. **Normal fulfillment**
   - Restaurant accepts order
   - Restaurant marks preparing
   - Restaurant marks ready
   - Courier claims delivery
   - Courier picks up
   - Courier delivers

### Scheduled Order Validation

- `scheduled_for` must be in the future
- `scheduled_for` must be within restaurant operating hours
- `scheduled_for` must be within region business hours
- Delivery window must be reasonable (configurable per region)
- Cancellation rules differ before vs after activation
- Timezone-aware validation using restaurant/region timezone

## Multi-Region Boundaries

### Region Data Model

- **Region**: Geographic area (country, market)
  - Has: code, name, country_code, timezone, currency_code, is_active
- **Region Config**: Platform settings per region
  - Has: platform fee, default delivery window, scheduled order lead time, delivery fee taxability
- **Tax Zone**: Tax jurisdiction within region
  - Has: code, name, country, state/province, city, postal code pattern
- **Tax Rate**: Tax rate per tax zone and category
  - Has: rate, inclusive/exclusive, applies to delivery fee, effective dates

### Region-Aware Behavior

1. **Restaurant belongs to region**
   - Restaurant has `region_id` foreign key
   - Restaurant currency derived from region
   - Restaurant timezone derived from region

2. **Order belongs to region**
   - Order has `region_id` (from restaurant)
   - Order has `tax_zone_id` (from delivery address)
   - Order currency snapshotted from restaurant/region
   - Order tax calculated using region/tax zone configuration

3. **Delivery scheduling is timezone-aware**
   - Restaurant opening hours stored in local time
   - Scheduled delivery times validated in local time
   - Worker activation uses UTC internally but converts to local time for validation

4. **Business hours are region-specific**
   - Each region has default business hours
   - Each restaurant can override business hours
   - Orders can only be placed during business hours (ASAP) or scheduled for business hours

### Region Isolation

- Queries scoped by `region_id` where appropriate
- List endpoints filter by user's region context
- Cross-region data access prevented by ownership checks
- Admin can view all regions, but operations are region-scoped
- Tax configuration is region-specific

## Region-Aware Scheduling

### Timezone Handling

- All timestamps stored in UTC in database
- Region timezone stored in `regions.timezone`
- Restaurant timezone derived from region (can override)
- Scheduled times validated in local time
- Worker activation compares UTC times but validates local business hours

### Scheduled Order Activation

```
Worker loop:
  1. Fetch scheduled orders where scheduled_for <= now (UTC)
  2. For each order:
     - Load region timezone
     - Convert scheduled_for to local time
     - Validate restaurant is open at local time
     - If valid: activate order (status: pending)
     - If invalid: mark as failed or reschedule
```

### Business Hour Validation

- Restaurant hours stored as day-of-week + open/close times in local time
- Validation converts UTC scheduled time to local time
- Checks if scheduled time falls within open hours
- Checks if scheduled time falls within region business hours

## Full CRUD Architecture Boundaries

### CRUD by Role

Each role has specific CRUD boundaries:

**Customer**
- Can CRUD own addresses
- Can read restaurants and menus
- Can create/read own orders
- Can cancel own orders (when allowed)
- Cannot access admin/owner/courier resources

**Restaurant Owner**
- Can CRUD own restaurants (soft delete)
- Can CRUD own restaurant hours
- Can CRUD own menu items (activate/deactivate, not hard delete if referenced)
- Can read/update workflow state for own orders
- Cannot access other owners' resources

**Courier**
- Can read own profile
- Can read available deliveries
- Can claim deliveries
- Can update pickup/complete states
- Cannot access customer/restaurant admin CRUD

**Admin**
- Full CRUD on regions, tax configuration
- Full CRUD on users (with suspend/disable, avoid hard delete)
- Full CRUD on restaurants (approve, suspend, archive)
- Read-only on audit logs
- Limited overrides on orders/payments (with audit)

### Ownership Enforcement

- Service layer checks ownership before allowing mutations
- Repository layer scopes queries by ownership where appropriate
- Handler layer validates role before calling service
- Middleware layer enforces coarse RBAC
- Object-level security prevents IDOR

### Workflow-State Constraints

- CRUD operations respect current workflow state
- Cannot edit orders in terminal states
- Cannot delete historical orders
- Cannot hard delete menu items referenced by orders
- Cannot modify tax rates used by historical orders
- Admin overrides create audit logs

## Future Scaling Notes

### Service Extraction

The modular monolith is designed for future service extraction:

- Clear domain boundaries make extraction straightforward
- Shared platform layer can be duplicated or extracted
- Outbox pattern enables eventual consistency between services
- Database can be split by domain when needed

### Horizontal Scaling

- API server is stateless (can scale horizontally)
- Worker processes can scale horizontally (SKIP LOCKED ensures safety)
- Database can use read replicas for read-heavy workloads
- Connection pooling limits prevent overwhelming database

### Caching Layer

- Add Redis or in-memory cache for:
  - Restaurant/menu data (cache invalidation on updates)
  - Region configuration (long TTL)
  - Tax configuration (long TTL)
  - User sessions (if needed)

### Message Broker

- Replace polling-based workers with message broker (RabbitMQ, Kafka)
- Outbox events published to message broker
- Workers consume from queues
- Better for high-volume scenarios

### Multi-Region Deployment

- Deploy separate database per region for latency
- Use read replicas for cross-region queries
- Region-aware routing in API gateway
- Eventual consistency between regions via replication

### Search Indexing

- Add Elasticsearch or similar for:
  - Restaurant search
  - Menu item search
  - Order search
- Outbox events trigger index updates

### Geospatial Search

- Add PostGIS for:
  - Restaurant location search
  - Delivery radius calculations
  - Courier location tracking

### Real-Time Features

- Add WebSocket support for:
  - Live order status updates
  - Courier location tracking
  - Restaurant order notifications

## Summary

This architecture provides a solid foundation for a production food delivery platform. The modular monolith design balances simplicity with future extensibility. Clear boundaries between domains, strict workflow enforcement, multi-region awareness, and comprehensive auditability make this suitable for real-world deployment.
