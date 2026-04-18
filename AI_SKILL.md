# AI_SKILL.md - Food Delivery API Backend

## Project Purpose

This is a production-style Food Delivery API Backend inspired by Uber Eats/DoorDash patterns. It is a modular monolith built with Go and PostgreSQL, designed for serious production use with strong security, strict workflow enforcement, multi-region support, tax-aware pricing, and comprehensive auditability.

**This is NOT a tutorial project.** It is a baseline for real-world deployment with operational readiness, scalability consciousness, and enterprise-grade security.

## Core Principles

1. **Security First**: Never weaken security, authorization, or auditability
2. **State-Driven Workflows**: All business flows must happen in strict sequences
3. **Multi-Region Ready**: Architecture supports multiple countries/markets cleanly
4. **Tax-Aware**: Server-side tax computation with data-driven configuration
5. **Audit Everything**: Sensitive actions create audit logs
6. **Idempotent**: Critical endpoints support idempotency
7. **No Secrets in Logs**: Never log passwords, tokens, or sensitive data
8. **Integer Money**: All monetary values stored in minor units (integers)
9. **UTC Timestamps**: All timestamps stored in UTC, preserve timezone context
10. **Role-Scoped CRUD**: Full CRUD access enforced by role, ownership, region, and workflow state

## Coding Standards

### Go Conventions

- Use Go 1.25+ syntax and idioms
- Follow standard Go project layout
- Use `context.Context` correctly throughout call chains
- Prefer clear code over clever code
- No global mutable state
- Use explicit error types for domain errors
- Use structured logging with redaction
- Handle errors explicitly, never ignore them

### Architecture Layers

```
Handler (thin) → Service (business logic) → Repository (SQL/data access)
```

- **Handlers**: HTTP request/response handling, validation, authorization checks
- **Services**: Business rules, workflow enforcement, transaction orchestration
- **Repositories**: SQL queries, data access, database-specific logic
- **DTOs**: Explicit request/response structs, never expose DB models directly

### File Organization

- Each domain module has: handler.go, service.go, repository.go, dto.go
- Platform modules in `internal/platform/` for shared infrastructure
- Clear separation between domain logic and infrastructure
- No circular dependencies

### Naming Conventions

- Files: lowercase with underscores (e.g., `user_service.go`)
- Functions: PascalCase for exported, camelCase for unexported
- Variables: camelCase
- Constants: PascalCase for exported, camelCase for unexported
- Database tables: snake_case
- Database columns: snake_case
- JSON tags: snake_case

## Architecture Rules

### Modular Monolith

- Clear boundaries between domains (auth, users, restaurants, menus, orders, delivery, payments, etc.)
- Each domain is self-contained with its own handlers, services, repositories
- Platform layer provides shared infrastructure (config, db, logger, middleware, etc.)
- Designed for future service extraction if needed

### Transaction Boundaries

- Critical writes (order creation, payment updates, etc.) must use transactions
- Transactions commit before outbox events are processed
- Workers only process committed and due work
- Use SKIP LOCKED for safe concurrent job claiming

### Outbox Pattern

- All async side effects use outbox events
- Outbox events created within the same transaction as the triggering change
- Workers poll for unprocessed outbox events
- Workers process events idempotently
- Failed events retry with exponential backoff
- Max retries before dead-letter marking

### State Machines

- Order statuses: pending_payment, scheduled, pending, accepted, rejected, preparing, ready_for_pickup, picked_up, delivered, cancelled
- Delivery statuses: pending_schedule, unassigned, assigned, picked_up, delivered, failed, cancelled
- Explicit transition validation in service layer
- Invalid transitions return domain errors
- Transition rules centralized, not scattered

## Migration Rules

### Database Migrations

- Use golang-migrate or goose
- Migrations are versioned and ordered
- Never modify existing migrations once deployed
- Use `up.sql` and `down.sql` for reversible changes
- Test migrations in development before production

### Schema Changes

- Additive changes preferred (new columns, new tables)
- Destructive changes require careful planning
- Use CHECK constraints for data validation
- Use foreign keys for referential integrity
- Index columns used in WHERE, JOIN, ORDER BY clauses
- Use partial indexes where appropriate

### Data Integrity

- UUID primary keys for all tables
- Timestamps (created_at, updated_at) on all tables
- Soft delete using `deleted_at` or `archived_at` for business entities
- No hard delete of historical orders, payments, deliveries, audit logs
- Tax rates, regions, restaurants use soft delete/archive

## API Response Conventions

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
    "details": { ... }
  },
  "request_id": "uuid"
}
```

### HTTP Status Codes

- 200: Success
- 201: Created
- 400: Bad Request (validation errors)
- 401: Unauthorized (no auth or invalid token)
- 403: Forbidden (auth but insufficient permissions)
- 404: Not Found
- 409: Conflict (idempotency, state conflicts)
- 422: Unprocessable Entity (business rule violations)
- 429: Too Many Requests (rate limit)
- 500: Internal Server Error

## Security Requirements

### Authentication

- Password hashing with bcrypt (cost 12) or argon2
- JWT access tokens with short TTL (15 minutes)
- Refresh tokens stored in database
- Refresh token rotation on every use
- Refresh token revocation support
- Logout endpoint invalidates refresh tokens
- Optional email verification hooks
- Optional password reset scaffolding

### Authorization

- Role-based access control (RBAC) middleware
- Roles: customer, restaurant_owner, courier, admin
- Ownership checks for restaurant owner resources
- Ownership checks for customer order access
- Courier-only actions restricted properly
- Admin-only actions enforced consistently
- Object-level security on every CRUD endpoint
- Prevent IDOR (Insecure Direct Object Reference)

### API Security

- Request validation on all write endpoints
- Strict DTOs, no blind binding to DB models
- Consistent error handling without leaking internals
- Rate limiting middleware (by IP and optionally user)
- Trusted CORS allowlist config
- Secure headers middleware
- Request size limits
- Timeout middleware
- Idempotency-key support for critical endpoints
- Input sanitization/normalization

### Data Security

- Never store raw passwords
- Never store raw card data (use payment provider tokenization)
- Never log secrets, tokens, or password hashes
- Redact sensitive fields in logs
- Minimize PII exposure in responses
- Separate internal vs public response DTOs
- Use environment variables for secrets
- Document secret rotation strategy
- Document encryption at rest/TLS expectations

### Audit/Compliance

- Audit log table for sensitive actions
- Actions: admin approvals, restaurant approval, role changes, order status overrides, payment status changes, tax configuration changes, region configuration changes
- Store: actor_id, actor_role, action, entity_type, entity_id, metadata, created_at
- Audit logs are immutable (no delete/update)

### Payment Security

- Real PCI-sensitive card storage is out of scope
- Delegate to external payment provider
- Use payment provider tokenization abstraction only
- Never persist raw card data
- Provide mock provider implementation for MVP

### Abuse/Fraud Baseline

- Rate limiting by IP and optionally user
- Idempotency for order creation/payment initiation
- Duplicate scheduled order protection
- Suspicious action logging hooks
- Document where anti-fraud systems would plug in

## Worker Rules

### Worker Types

1. **Outbox Event Processor**: Processes async side effects
2. **Scheduled Order Activator**: Activates scheduled orders when due
3. **Delayed Notification Sender**: Sends delayed notifications
4. **Payment Reconciler**: Follows up on payments for mock provider
5. **Maintenance Jobs**: Cleanup, maintenance tasks

### Worker Design

- Polling-based worker is acceptable
- DB-backed jobs/outbox table
- Safe claiming with row locking and SKIP LOCKED
- Fields: retry_count, max_retries, next_attempt_at, locked_at, locked_by, processed_at, failed_at, last_error
- Workers must be idempotent
- Workers must support graceful shutdown
- Workers must log structured events
- Workers must only process committed and due work

### Retry Strategy

- Exponential backoff for retries
- Max retries before dead-letter
- Mark failed jobs with last_error
- Dead-letter jobs require manual inspection
- Cleanup old processed jobs periodically

### Worker Safety

- Never process uncommitted transactions
- Validate job state before processing
- Handle duplicate job processing gracefully
- Log all job processing attempts
- Support worker scaling (multiple workers safe)

## State Transition Rules

### Order Status Transitions

Allowed transitions:
- pending_payment → scheduled (if scheduled_for is future)
- pending_payment → pending (if ASAP)
- scheduled → pending (when activated by worker)
- pending → accepted
- pending → rejected
- accepted → preparing
- accepted → cancelled
- rejected → (terminal)
- preparing → ready_for_pickup
- preparing → cancelled
- ready_for_pickup → picked_up
- ready_for_pickup → cancelled
- picked_up → delivered
- picked_up → failed
- delivered → (terminal)
- cancelled → (terminal)

Invalid transitions must return domain errors.

### Delivery Status Transitions

Allowed transitions:
- pending_schedule → unassigned (when order is pending/accepted)
- unassigned → assigned (when courier claims)
- assigned → picked_up
- assigned → cancelled
- picked_up → delivered
- picked_up → failed
- delivered → (terminal)
- failed → (terminal)
- cancelled → (terminal)

### Transition Guards

- Order cannot be marked preparing before accepted
- Order cannot be marked ready before preparing
- Courier cannot pick up before claiming/assignment
- Courier cannot complete delivery before pickup
- Scheduled orders must not enter active fulfillment before activation window
- Async actions triggered by flow step only after DB transaction commit

## Multi-Region Rules

### Region Awareness

- Restaurants belong to a region
- Users/addresses resolve to a region
- Orders created within a region context
- Delivery scheduling is timezone-aware per region
- Taxes computed using region/tax zone attached to order
- All critical monetary/tax values snapshotted on order at creation
- Region config is data-driven, not hardcoded
- Future region expansion should not require major rewrites

### Timezone Handling

- Use UTC for internal timestamp storage
- Preserve region timezone context for:
  - Restaurant opening hours
  - Scheduled delivery activation
  - Customer-facing schedule validation
  - Regional business day logic
- Never rely on server local time
- Convert to region timezone only for display/validation

### Currency Handling

- Each region has a default currency
- Each restaurant operates in exactly one currency
- Each order snapshots currency_code at creation
- Each payment tied to order currency
- Avoid FX conversion inside MVP
- Document future extension points for FX/conversion

## Tax Rules

### Tax Computation

- Never trust tax values from client
- Determine region/tax zone from restaurant + delivery address
- Fetch applicable tax configuration from DB
- Compute line-item tax
- Compute delivery-fee tax if applicable
- Compute total tax
- Store tax snapshot on order and order_items
- Preserve enough detail for invoice/receipt generation

### Tax Configuration

- Tax zones: geographic areas with specific tax rules
- Tax rates: specific rates per zone and category
- Tax categories: product categories for tax classification
- Support inclusive vs exclusive tax models
- Support VAT/GST/sales-tax styles
- Item-level taxability
- Delivery-fee taxability configurable
- Tax-inclusive pricing in some regions
- Tax-exclusive pricing in some regions
- Tax rounding rules consistent

### Tax Snapshot Strategy

- Order tax lines stored in separate table or JSONB
- Tax changes do not retroactively alter existing orders
- Tax-inclusive vs exclusive mode snapshotted per order
- Tax rate snapshotted at order creation time
- Preserve tax breakdown for audit/invoice purposes

### Tax Calculation Service

- Centralized in dedicated tax service/module
- Not scattered across handlers
- Server-side only
- Data-driven from database
- Support both inclusive and exclusive modes
- Configurable delivery fee taxation

## Money Handling Rules

### Integer Minor Units

- Store money in minor units using integer columns
- Never use float for money
- Never use imprecise client-side totals
- Every monetary value includes currency_code
- Subtotal, delivery fee, tax, total all in minor units

### Currency Fields

- currency_code on orders, payments, restaurants
- Currency code snapshotted at order creation
- Display formatting handled client-side
- Server always stores in minor units

### Price Snapshots

- Order items snapshot: item_name, unit_price_amount, line_subtotal_amount, line_tax_amount, line_total_amount, currency_code
- Order snapshots: subtotal_amount, delivery_fee_amount, tax_amount, total_amount, currency_code
- Historical orders never re-priced
- Menu price changes do not affect existing orders

## Role-Scoped CRUD Rules

### Customer

- users/me: read/update own profile
- user_addresses: full CRUD on own addresses
- restaurants: read-only
- menus: read-only
- orders: create, read own, list own, cancel own when allowed
- payments: read own payment status, no direct delete
- No access to admin/owner-only resources
- No access to other customers' data

### Restaurant Owner

- Own restaurant: create, read, update
- Own restaurant deletion: soft-delete/archive only
- Own restaurant hours: full CRUD
- Own menu items: full CRUD
- Own restaurant orders: read/list and workflow updates
- Cannot delete historical orders
- Cannot manage restaurants owned by others
- Can activate/deactivate menu items (not hard delete when referenced historically)

### Courier

- Own profile: read/update limited courier profile fields
- Delivery jobs: read available, claim, read own assigned deliveries, update pickup/complete states
- No arbitrary delete of delivery history
- No access to customer or restaurant admin CRUD outside role scope

### Admin

- Users: full admin read/list, limited update, controlled disable/suspend, avoid destructive hard delete
- Restaurants: full admin read/list/approve/update/suspend/archive
- Regions: full CRUD
- Region configs: full CRUD
- Tax categories: full CRUD
- Tax zones: full CRUD
- Tax rates: full CRUD
- Audit logs: read-only
- Orders: read/list and limited administrative override endpoints with audit logs
- Payments: read/list/update status overrides with audit logs

### System/Internal

- Jobs/outbox/audit/idempotency tables not exposed as public general CRUD
- Expose admin inspection endpoints only where appropriate
- Do not expose dangerous raw internal CRUD for infrastructure tables

## Soft Delete/Archive Rules

### Soft Delete Required For

- restaurants
- menu_items
- users (disable/suspend instead of hard delete)
- tax_rates (deactivate instead of hard delete once used)
- regions (deactivate instead of hard delete once referenced)

### Hard Delete Allowed For

- restaurant_hours (if not historically relevant)
- addresses (by owner if not referenced by active orders)

### Never Hard Delete

- orders
- payments
- deliveries
- audit_logs
- outbox/job history

### Soft Delete Pattern

Use one consistent pattern:
- `is_active` boolean
- `archived_at` timestamp
- `deleted_at` timestamp
- `suspended_at` timestamp

Choose one and document it. This project uses `deleted_at` for soft delete.

## Business-State Constraints on CRUD

### Menu Items

- Menu items in historical orders may be archived/inactivated but not hard-deleted
- Active menu items can be ordered
- Inactive menu items cannot be ordered

### Orders

- Accepted/preparing/ready/picked-up/delivered orders cannot be edited like draft resources
- Customers may cancel only when cancellation rules permit
- Restaurants cannot arbitrarily rewrite historical order pricing after creation
- Admin overrides must create audit logs

### Tax Rates

- Tax rates used by historical orders may be deactivated for future use but must not mutate past order snapshots

### Scheduled Orders

- May only be edited/cancelled within allowed pre-activation rules
- Once activated, normal cancellation rules apply

### Delivery Assignments

- Cannot be deleted once fulfillment has started
- Status changes follow strict transition rules

## What AI Assistants Must Do

When modifying this project:

1. **Read AI_SKILL.md first** - Understand the project principles before making changes
2. **Maintain security posture** - Never weaken authentication, authorization, or auditability
3. **Preserve workflow enforcement** - State transitions must remain strict and validated
4. **Keep tax server-side** - Never trust client-submitted tax values
5. **Use integer money** - Never use float for monetary values
6. **Store UTC timestamps** - Always use UTC, preserve timezone context
7. **Enforce RBAC** - Every endpoint must check roles and ownership
8. **Log securely** - Never log secrets, tokens, password hashes, or sensitive payloads
9. **Use transactions** - Critical writes must be transactional
10. **Maintain idempotency** - Critical endpoints must support idempotency
11. **Create audit logs** - Sensitive admin/business actions must create audit logs
12. **Soft delete appropriately** - Use soft delete for business entities, never hard delete historical data
13. **Test changes** - Ensure code compiles and tests pass
14. **Update documentation** - Keep docs in sync with code changes
15. **Follow naming conventions** - Maintain consistent naming throughout

## What AI Assistants Must NOT Do

When modifying this project:

1. **Do NOT weaken security** - Never bypass auth, RBAC, ownership, or audit checks
2. **Do NOT bypass idempotency** - Never remove idempotency from critical endpoints
3. **Do NOT bypass workflow** - Never allow invalid state transitions
4. **Do NOT bypass tax rules** - Never hardcode tax percentages or trust client tax values
5. **Do NOT bypass scheduling** - Never allow scheduled orders to activate early
6. **Do NOT use float for money** - Never introduce floating-point money handling
7. **Do NOT log secrets** - Never add logging of passwords, tokens, or sensitive data
8. **Do NOT hard delete historical data** - Never allow deletion of orders, payments, deliveries, audit logs
9. **Do NOT skip ownership checks** - Never allow cross-owner data access
10. **Do NOT skip region scope** - Never allow cross-region data leakage
11. **Do NOT introduce global state** - Never use global mutable variables
12. **Do NOT use fake repositories** - Never replace real DB repositories with in-memory fakes
13. **Do NOT ignore errors** - Never silently ignore errors
14. **Do NOT break transactions** - Never remove transaction boundaries from critical writes
15. **Do NOT break worker safety** - Never allow workers to process uncommitted or non-due work

## Summary

This project is a serious production baseline. Every change must maintain security, correctness, auditability, and operational readiness. When in doubt, choose the more conservative, secure option. The goal is a backend that can safely handle real users, real money, and real business operations.

**If you are unsure whether a change is safe, do not make it. Ask for clarification first.**
