# Design Decisions

## Overview

This document explains the key design decisions, tradeoffs, and rationale behind the Food Delivery API Backend architecture.

## Why Modular Monolith

### Decision

We chose a modular monolith architecture over microservices for the initial implementation.

### Rationale

1. **Simplicity**: Single codebase, single deployment, single database reduces operational complexity
2. **Development Velocity**: Faster iteration without distributed system overhead
3. **Transaction Simplicity**: ACID transactions across domains without distributed transaction complexity
4. **Testing**: Easier to write integration tests with a single process
5. **Cost**: Lower infrastructure cost for initial deployment
6. **Team Size**: Suitable for small to medium teams

### Tradeoffs

**Pros**:
- Simpler deployment and monitoring
- No network latency between services
- Shared database simplifies queries
- Easier debugging and tracing
- Lower operational overhead

**Cons**:
- Harder to scale individual components independently
- Technology choices shared across entire system
- Larger blast radius for failures
- Potential for tighter coupling between modules

### Mitigation

- Clear module boundaries make future service extraction straightforward
- Outbox pattern enables eventual consistency if services are split
- Platform layer can be extracted independently
- Database can be partitioned by domain when needed

## Why Go + PostgreSQL

### Go Language Choice

**Rationale**:
1. **Performance**: Compiled language, fast execution, low memory footprint
2. **Concurrency**: Goroutines and channels make concurrent programming simple
3. **Type Safety**: Strong typing catches errors at compile time
4. **Standard Library**: Rich standard library reduces dependencies
5. **Deployment**: Single binary deployment simplifies operations
6. **Ecosystem**: Mature ecosystem for web services, databases, and tooling
7. **Developer Experience**: Clear syntax, good tooling, fast compilation

**Tradeoffs**:
- Steeper learning curve than Python/Ruby for some developers
- More verbose than dynamic languages
- Generics (added in Go 1.18) are still maturing

### PostgreSQL Choice

**Rationale**:
1. **ACID Compliance**: Strong transaction guarantees
2. **JSONB Support**: Flexible schema for metadata and snapshots
3. **Full-Text Search**: Built-in search capabilities
4. **Extensions**: PostGIS for geospatial, pgcrypto for encryption
5. **Maturity**: Battle-tested, reliable, well-documented
6. **Community**: Large community, good tooling
7. **Performance**: Excellent performance for read/write workloads

**Tradeoffs**:
- Vertical scaling limits compared to distributed databases
- Requires manual sharding for horizontal scaling at very large scale
- Higher operational complexity than managed NoSQL services

### pgx/pgxpool Choice

**Rationale**:
1. **Performance**: Faster than database/sql driver
2. **Feature Support**: Better support for PostgreSQL-specific features
3. **Connection Pooling**: Built-in connection pooling with pgxpool
4. **Context Support**: Proper context support for cancellation
5. **Type Safety**: Better type mapping for PostgreSQL types

## Why Outbox Pattern

### Decision

We use the outbox pattern for reliable async side effects instead of direct async calls or message queues.

### Rationale

1. **Reliability**: Outbox events created in same transaction as triggering change
2. **Atomicity**: Either both the change and event commit, or neither does
3. **No Data Loss**: Events persisted even if worker is down
4. **Idempotency**: Workers can safely retry processing
5. **Audit Trail**: Outbox table provides audit trail of events
6. **Simple Deployment**: No additional infrastructure required initially

### Tradeoffs

**Pros**:
- No message broker infrastructure needed initially
- Exactly-once semantics with transactional outbox
- Simple to implement and understand
- Works with polling-based workers

**Cons**:
- Polling has higher latency than push-based messaging
- Database load from polling queries
- Requires careful job claiming logic (SKIP LOCKED)
- Not suitable for very high throughput scenarios

### Mitigation

- Can migrate to message broker (RabbitMQ, Kafka) when needed
- Outbox table can be used as source for event streaming
- Polling interval can be tuned based on requirements
- SKIP LOCKED ensures safe concurrent processing

## Why Worker-Based Async Design

### Decision

We use background workers for async processing instead of goroutines or inline execution.

### Rationale**

1. **Reliability**: Workers can retry failed operations
2. **Observability**: Worker processing is logged and monitored
3. **Resource Control**: Workers can be scaled independently
4. **Graceful Shutdown**: Workers can drain queues before shutdown
5. **Separation of Concerns**: API handlers remain fast, workers handle slow operations
6. **Idempotency**: Workers designed for idempotent processing

### Tradeoffs

**Pros**:
- API responses remain fast
- Failed operations can retry automatically
- Workers can be scaled independently
- Better resource utilization

**Cons**:
- Eventual consistency for async operations
- Additional operational complexity
- Requires job queue infrastructure

### Mitigation**

- Outbox pattern ensures no lost events
- Workers process events quickly (typically < 1 second)
- Status endpoints allow clients to poll for async operation completion
- Webhooks can notify clients of completion (future enhancement)

## Why Region-Aware Design

### Decision

We built region-awareness into the core data model and business logic.

### Rationale**

1. **Multi-Market Support**: Platform can operate in multiple countries/markets
2. **Compliance**: Different tax rules per region
3. **Currency**: Different currencies per region
4. **Timezones**: Business hours and scheduling per region timezone
5. **Configuration**: Platform settings can vary per region
6. **Future Expansion**: Adding new regions requires minimal code changes

### Tradeoffs

**Pros**:
- Clean separation of regional concerns
- Easy to add new regions
- Regional configuration is data-driven
- Compliance requirements met

**Cons**:
- Additional complexity in data model
- More joins for region-related queries
- Need to handle timezone conversions

### Mitigation

- Region ID cached in user session for performance
- Regional configuration cached with long TTL
- Timezone conversions only where needed
- Database indexes on region_id for performance

## Why Integer Money Fields

### Decision

We store all monetary values as integers in minor units (cents, pence, etc.) instead of floats or decimals.

### Rationale**

1. **Precision**: No floating-point rounding errors
2. **Predictability**: Consistent behavior across all operations
3. **Performance**: Integer arithmetic is fast
4. **Simplicity**: No need for decimal libraries
5. **Database Support**: All databases support integers well
6. **Auditability**: Exact values stored and audited

### Tradeoffs

**Pros**:
- No precision loss
- Simple arithmetic
- Fast operations
- Exact audit trail

**Cons**:
- Need to track currency separately
- Display formatting required client-side
- Division operations require care (rounding)

### Mitigation

- Currency_code stored alongside all monetary values
- Rounding rules documented and implemented consistently
- Display formatting handled client-side or in service layer
- Division operations use banker's rounding or documented rules

## Tax Design Tradeoffs

### Decision

We implemented a flexible, data-driven tax system with support for inclusive/exclusive modes.

### Rationale**

1. **Flexibility**: Different regions have different tax models (VAT, GST, sales tax)
2. **Data-Driven**: Tax rules configurable without code changes
3. **Compliance**: Supports both tax-inclusive and tax-exclusive pricing
4. **Auditability**: Tax snapshots preserved for historical orders
5. **Future-Proof**: Can handle complex tax scenarios as needed

### Tradeoffs

**Pros**:
- Configurable without code changes
- Supports multiple tax models
- Tax changes don't affect historical orders
- Audit trail for tax calculations

**Cons**:
- More complex data model
- Additional joins for tax calculation
- Need to manage tax configuration lifecycle

### Mitigation

- Tax configuration cached with long TTL
- Tax calculation service centralizes logic
- Tax snapshots prevent retroactive changes
- Admin UI for tax configuration management

## Security Posture

### Decision

We implemented defense-in-depth security with multiple layers of protection.

### Rationale**

1. **Defense in Depth**: Multiple layers reduce attack surface
2. **Principle of Least Privilege**: Users only have access to what they need
3. **Auditability**: All sensitive actions logged
4. **Compliance**: Meets regulatory requirements for financial systems
5. **Trust**: Customers trust systems that take security seriously

### Tradeoffs

**Pros**:
- Strong security posture
- Compliance with best practices
- Audit trail for investigations
- Reduced risk of data breaches

**Cons**:
- Additional complexity in code
- Performance overhead from security checks
- More operational overhead

### Mitigation

- Security checks optimized with indexes
- Caching for frequently accessed data
- Security middleware is efficient
- Benefits outweigh costs for production systems

## Role-Scoped CRUD Tradeoffs

### Decision

We implemented strict role-scoped CRUD with ownership checks and workflow constraints.

### Rationale**

1. **Security**: Prevents unauthorized data access
2. **Compliance**: Meets data protection regulations
3. **Business Logic**: Enforces business rules at data access level
4. **Auditability**: All data access logged
5. **User Experience**: Users only see relevant data

### Tradeoffs

**Pros**:
- Strong data access control
- Prevents IDOR vulnerabilities
- Enforces business rules
- Audit trail for data access

**Cons**:
- Additional complexity in handlers and services
- More database queries for ownership checks
- Potential for over-restrictive access

### Mitigation

- Ownership checks optimized with indexes
- Repository layer scopes queries by ownership
- Clear documentation of access rules
- Regular security reviews

## Future Global Scaling Notes

### Service Extraction

When the system outgrows the modular monolith:

1. **Identify Boundaries**: Extract domains with clear boundaries (orders, payments, notifications)
2. **Define APIs**: Use gRPC or HTTP for inter-service communication
3. **Database Splitting**: Split database by domain with foreign key replacement
4. **Event Streaming**: Use message broker for eventual consistency
5. **Monitoring**: Add distributed tracing (OpenTelemetry)

### Multi-Region Deployment

For global deployment:

1. **Regional Databases**: Deploy separate database per region for latency
2. **Read Replicas**: Use read replicas for cross-region queries
3. **API Gateway**: Region-aware routing in API gateway
4. **Data Replication**: Async replication between regions for analytics
5. **Local Caching**: Regional cache layers for performance

### Caching Strategy

When cache layer is needed:

1. **Redis**: Distributed cache for session data, rate limiting
2. **Application Cache**: In-memory cache for configuration, tax rules
3. **CDN**: Cache static assets and API responses where appropriate
4. **Cache Invalidation**: Outbox events trigger cache invalidation

### Search Integration

When search is needed:

1. **Elasticsearch**: Full-text search for restaurants, menu items
2. **PostgreSQL Full-Text**: Simple search for initial implementation
3. **Search Indexing**: Outbox events trigger index updates
4. **Geospatial Search**: PostGIS for location-based search

### Real-Time Features

When real-time features are needed:

1. **WebSockets**: Live order status updates
2. **Server-Sent Events**: One-way updates to clients
3. **Push Notifications**: Mobile push for order updates
4. **Courier Tracking**: Real-time location updates

## Concurrency Strategy

### Courier Claim Concurrency

We use database-level locking to prevent double-claim of deliveries:

```sql
SELECT * FROM deliveries
WHERE id = $1 AND status = 'unassigned'
FOR UPDATE SKIP LOCKED;
```

**Rationale**:
- Database guarantees atomicity
- SKIP LOCKED prevents blocking
- No need for distributed locks
- Simple and reliable

### Order Creation Concurrency

Order creation uses transactions to prevent race conditions:

1. Begin transaction
2. Lock restaurant row (if needed)
3. Validate menu item availability
4. Create order and related records
5. Commit transaction

**Rationale**:
- ACID guarantees prevent partial updates
- Row locks prevent concurrent modifications
- Simple and reliable

### Worker Concurrency

Workers use SKIP LOCKED for safe concurrent processing:

```sql
SELECT * FROM jobs
WHERE status = 'pending' AND next_attempt_at <= now()
FOR UPDATE SKIP LOCKED
LIMIT 100;
```

**Rationale**:
- Multiple workers can run safely
- No double processing
- Simple to implement
- Database handles locking

## Scheduling Strategy

### Scheduled Order Activation

Scheduled orders activated by worker:

1. Worker polls for scheduled orders where `scheduled_for <= now`
2. For each order, validates restaurant is still open
3. Activates order (status: pending)
4. Creates outbox event for activation

**Rationale**:
- Decouples activation from order creation
- Workers can retry failed activations
- Restaurant hours can change between scheduling and activation
- Audit trail for activations

### Timezone Handling

Timezone conversions handled at service layer:

1. Store all timestamps in UTC in database
2. Store region timezone in regions table
3. Convert to local time for validation and display
4. Convert back to UTC for storage

**Rationale**:
- UTC storage is unambiguous
- Local time for user-facing operations
- Timezone changes don't affect stored data
- Standard practice for multi-timezone systems

## Summary

These design decisions prioritize reliability, security, and maintainability over premature optimization. The system is designed to scale when needed, but starts simple enough for rapid development and deployment. The modular architecture allows for incremental evolution as requirements grow.
