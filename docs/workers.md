# Workers

## Overview

This document describes the background worker system for the Food Delivery API Backend. Workers handle asynchronous tasks, scheduled order activation, outbox event processing, and maintenance jobs.

## Worker Types

### 1. Outbox Event Processor

Processes outbox events created during transaction commits.

**Purpose**: Reliable async side effects without message broker

**Events Handled**:
- `order_created`: Send notifications to restaurant and customer
- `order_status_changed`: Send status update notifications
- `payment_completed`: Send payment confirmation
- `delivery_assigned`: Notify courier of assignment
- `delivery_picked_up`: Notify customer of pickup
- `delivery_delivered`: Send delivery confirmation
- `restaurant_approved`: Notify restaurant owner of approval
- `tax_config_changed`: Clear relevant caches

**Processing Logic**:
1. Poll for unprocessed outbox events
2. Lock rows with `SELECT FOR UPDATE SKIP LOCKED`
3. Dispatch to event handler based on event type
4. Mark event as processed on success
5. Retry on failure with exponential backoff
6. Mark as dead-letter after max retries

### 2. Scheduled Order Activator

Activates scheduled orders when their scheduled time arrives.

**Purpose**: Transition scheduled orders to active fulfillment pipeline

**Processing Logic**:
1. Poll for scheduled orders where `scheduled_for <= now()` and status = `scheduled`
2. For each order:
   - Load restaurant and region timezone
   - Convert scheduled_for to local time
   - Validate restaurant is still open at local time
   - If valid: activate order (status = `pending`, set `activated_at`)
   - If invalid: mark as failed or reschedule
   - Create outbox event for activation
3. Create notification job for customer and restaurant

**Validation Rules**:
- Restaurant must still be active and approved
- Restaurant must be open at scheduled time (local timezone)
- Order must not already be cancelled
- Region must still allow scheduled orders

### 3. Delayed Notification Sender

Sends delayed notifications at scheduled times.

**Purpose**: Send reminders and follow-up notifications

**Notification Types**:
- Order reminder (15 minutes before scheduled delivery)
- Delivery ETA updates
- Courier pickup reminders
- Restaurant preparation reminders

**Processing Logic**:
1. Poll for notification jobs where `next_attempt_at <= now()` and status = `pending`
2. Lock rows with `SELECT FOR UPDATE SKIP LOCKED`
3. Send notification via notification service
4. Mark job as succeeded on success
5. Retry on failure with exponential backoff

### 4. Payment Reconciler

Reconciles payment status with payment provider.

**Purpose**: Ensure payment status is accurate

**Processing Logic**:
1. Poll for payments with status = `processing` where `created_at > now() - 1 hour`
2. Query payment provider for status
3. Update payment status based on provider response
4. If succeeded: trigger order status transition
5. If failed: mark order as failed and notify customer
6. Create outbox event for status change

**Provider Abstraction**:
- Mock provider for development/testing
- Stripe integration for production (future)
- Provider-specific logic isolated in provider interface

### 5. Maintenance Jobs

Performs periodic maintenance tasks.

**Job Types**:
- Cleanup old processed outbox events (older than 30 days)
- Cleanup old audit logs (if retention policy configured)
- Refresh materialized views (if any)
- Archive old completed jobs (older than 90 days)
- Generate daily reports (if configured)

**Processing Logic**:
1. Run on schedule (e.g., daily at 2 AM)
2. Execute maintenance tasks in sequence
3. Log results and metrics
4. Alert on failures

## Polling Strategy

### Polling Interval

Default polling intervals:
- Outbox Event Processor: 5 seconds
- Scheduled Order Activator: 10 seconds
- Delayed Notification Sender: 30 seconds
- Payment Reconciler: 60 seconds
- Maintenance Jobs: Daily

### Polling Query

General pattern for polling:

```sql
SELECT * FROM jobs
WHERE status IN ('pending', 'failed')
  AND next_attempt_at <= now()
ORDER BY next_attempt_at ASC
FOR UPDATE SKIP LOCKED
LIMIT 100;
```

**SKIP LOCKED** ensures:
- Multiple workers can run safely
- No double processing
- Non-blocking concurrent access
- Simple implementation without message broker

### Batch Processing

Workers process jobs in batches:
- Default batch size: 100 jobs
- Configurable per worker type
- Process batch, then poll again
- Reduces database round trips

## Retry Strategy

### Exponential Backoff

Retry delays:
- Attempt 1: 1 second
- Attempt 2: 2 seconds
- Attempt 3: 4 seconds
- Attempt 4: 8 seconds
- Attempt 5: 16 seconds
- Max retry: 3-5 attempts (configurable)

**Formula**: `delay = base_delay * (2 ^ (retry_count - 1))`

### Max Retries

Default max retries per job type:
- Outbox events: 5 retries
- Notifications: 3 retries
- Payment reconciliation: 10 retries (more retries for financial operations)
- Scheduled order activation: 3 retries
- Maintenance jobs: 1 retry (manual intervention on failure)

### Dead-Letter Queue

Jobs that exceed max retries are marked as dead-letter:

```sql
UPDATE jobs
SET status = 'dead_letter',
    failed_at = now(),
    last_error = 'max retries exceeded'
WHERE id = $1;
```

Dead-letter jobs:
- Require manual inspection
- Can be retried manually via admin interface
- Alert on dead-letter accumulation

## Idempotency

### Worker Idempotency

Workers must be idempotent to handle duplicate processing:

**Strategies**:
1. Check if work already done before processing
2. Use unique constraints in database
3. Check entity status before applying changes
4. Use conditional updates (UPDATE ... WHERE status = X)

**Example - Order Activation**:
```sql
UPDATE orders
SET status = 'pending',
    activated_at = now()
WHERE id = $1
  AND status = 'scheduled';
```

If no rows updated, order was already activated or in different state.

### Outbox Event Idempotency

Outbox events marked as processed after successful handling:

```sql
UPDATE outbox_events
SET processed = true,
    processed_at = now()
WHERE id = $1
  AND processed = false;
```

If no rows updated, event was already processed by another worker.

## Failure Handling

### Transient Failures

Retryable failures:
- Network timeouts
- Database connection errors
- Temporary provider unavailability
- Rate limiting

**Action**: Retry with exponential backoff

### Permanent Failures

Non-retryable failures:
- Invalid data
- Business rule violations
- Entity not found
- Permission errors

**Action**: Mark as failed, do not retry, alert for investigation

### Failure Logging

All failures logged with:
- Job ID and type
- Error message
- Stack trace
- Retry count
- Payload (redacted if sensitive)

### Alerting

Alert on:
- Dead-letter queue accumulation (> 100 jobs)
- High failure rate (> 10% failure rate in last hour)
- Worker crash/downtime
- Scheduled order activation failures

## Schedule Activation Flow

### Detailed Flow

```
1. Worker polls for scheduled orders:
   SELECT * FROM orders
   WHERE status = 'scheduled'
     AND scheduled_for <= now()
   FOR UPDATE SKIP LOCKED
   LIMIT 50;

2. For each order:
   a. Load restaurant (owner_id, region_id, timezone)
   b. Load region (timezone, config)
   c. Load restaurant hours for day
   d. Convert scheduled_for to local time
   e. Validate:
      - Restaurant is_active = true
      - Restaurant is_approved = true
      - Restaurant is open at local time
      - Region allow_scheduled_orders = true
   f. If valid:
      BEGIN TRANSACTION
        UPDATE orders
        SET status = 'pending',
            activated_at = now()
        WHERE id = order_id;
        
        UPDATE deliveries
        SET status = 'unassigned'
        WHERE order_id = order_id;
        
        INSERT INTO outbox_events (event_type, aggregate_type, aggregate_id, payload)
        VALUES ('order_activated', 'order', order_id, '{...}');
      COMMIT
   g. If invalid:
      UPDATE orders
      SET status = 'cancelled',
          cancellation_reason = 'Restaurant closed at scheduled time',
          cancelled_at = now()
      WHERE id = order_id;
      
      INSERT INTO outbox_events (...)
      VALUES ('order_cancelled', ...);

3. Log processing result
4. Mark outbox events for notifications
```

### Timezone Handling

```go
// Convert UTC scheduled time to restaurant local time
localTime := scheduledFor.In(timeZone)

// Check if restaurant is open
isOpen := restaurantService.IsOpen(restaurantID, localTime)

// If not open, cancel order
if !isOpen {
    cancelOrder(orderID, "Restaurant closed at scheduled time")
}
```

## Outbox Flow

### Event Creation

Events created in same transaction as triggering change:

```go
func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error) {
    tx, err := s.db.Begin(ctx)
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    // Create order
    order, err := s.createOrderInTx(ctx, tx, req)
    if err != nil {
        return nil, err
    }

    // Create outbox event
    event := &OutboxEvent{
        EventType:     "order_created",
        AggregateType: "order",
        AggregateID:   order.ID,
        Payload:       order.ToPayload(),
    }
    if err := s.outboxRepo.CreateInTx(ctx, tx, event); err != nil {
        return nil, err
    }

    if err := tx.Commit(); err != nil {
        return nil, err
    }

    return order, nil
}
```

### Event Processing

```go
func (w *OutboxWorker) processEvent(ctx context.Context, event *OutboxEvent) error {
    switch event.EventType {
    case "order_created":
        return w.handleOrderCreated(ctx, event)
    case "order_status_changed":
        return w.handleOrderStatusChanged(ctx, event)
    case "payment_completed":
        return w.handlePaymentCompleted(ctx, event)
    default:
        log.Warn("Unknown event type", "type", event.EventType)
        return nil
    }
}

func (w *OutboxWorker) handleOrderCreated(ctx context.Context, event *OutboxEvent) error {
    var payload OrderPayload
    if err := json.Unmarshal(event.Payload, &payload); err != nil {
        return err
    }

    // Send notification to restaurant
    if err := w.notificationService.SendRestaurantNotification(ctx, payload.RestaurantID, payload); err != nil {
        return err
    }

    // Send notification to customer
    if err := w.notificationService.SendCustomerNotification(ctx, payload.CustomerID, payload); err != nil {
        return err
    }

    return nil
}
```

### Event Marking

```go
func (w *OutboxWorker) markProcessed(ctx context.Context, eventID uuid.UUID) error {
    query := `
        UPDATE outbox_events
        SET processed = true,
            processed_at = now()
        WHERE id = $1
          AND processed = false
    `
    result, err := w.db.Exec(ctx, query, eventID)
    if err != nil {
        return err
    }
    if result.RowsAffected() == 0 {
        // Event already processed by another worker
        return nil
    }
    return nil
}
```

## Due-Job Processing Rules

### Job Selection

Workers only process jobs that are:
1. Status = `pending` or `failed`
2. `next_attempt_at <= now()`
3. Not already locked (SKIP LOCKED)

### Job Locking

```sql
SELECT * FROM jobs
WHERE status IN ('pending', 'failed')
  AND next_attempt_at <= now()
ORDER BY next_attempt_at ASC
FOR UPDATE SKIP LOCKED
LIMIT 100;
```

Update lock on selection:
```sql
UPDATE jobs
SET status = 'processing',
    locked_at = now(),
    locked_by = $1
WHERE id = ANY($2);
```

### Job Completion

On success:
```sql
UPDATE jobs
SET status = 'succeeded',
    processed_at = now()
WHERE id = $1;
```

On failure (retryable):
```sql
UPDATE jobs
SET status = 'failed',
    retry_count = retry_count + 1,
    next_attempt_at = now() + (2 ^ retry_count) * interval '1 second',
    last_error = $2,
    failed_at = now()
WHERE id = $1
  AND retry_count < max_retries;
```

On failure (max retries exceeded):
```sql
UPDATE jobs
SET status = 'dead_letter',
    failed_at = now(),
    last_error = $2
WHERE id = $1;
```

## Worker Safety

### Transaction Commit Requirement

Workers only process committed transactions:
- Outbox events only created after transaction commit
- Workers poll for events, not uncommitted data
- No race conditions between API and workers

### Job State Validation

Workers validate job state before processing:
```go
job, err := w.repo.GetForProcessing(ctx, jobID)
if err != nil {
    return err
}
if job.Status != "pending" && job.Status != "failed" {
    // Job already processed, skip
    return nil
}
```

### Duplicate Handling

Workers handle duplicate job processing gracefully:
- Check if work already done
- Use conditional updates
- Return success if work already completed

### Graceful Shutdown

Workers support graceful shutdown:
1. Stop polling for new jobs
2. Complete in-progress jobs
3. Close database connections
4. Exit cleanly

```go
func (w *Worker) Shutdown(ctx context.Context) error {
    w.stop()
    w.wg.Wait()
    return w.db.Close()
}
```

### Worker Scaling

Multiple workers can run safely:
- SKIP LOCKED prevents double processing
- Each worker processes different jobs
- No coordination needed between workers
- Horizontal scaling supported

## Worker Configuration

### Environment Variables

```
# Worker settings
WORKER_POLL_INTERVAL=5s
WORKER_BATCH_SIZE=100
WORKER_MAX_RETRIES=5
WORKER_RETRY_BASE_DELAY=1s

# Worker-specific settings
OUTBOX_WORKER_ENABLED=true
SCHEDULED_ACTIVATOR_WORKER_ENABLED=true
NOTIFICATION_WORKER_ENABLED=true
PAYMENT_RECONCILER_WORKER_ENABLED=true
MAINTENANCE_WORKER_ENABLED=true
```

### Concurrency Settings

```
# Worker goroutines
WORKER_CONCURRENCY=10

# Job processing
WORKER_JOB_TIMEOUT=30s
WORKER_JOB_HEARTBEAT=10s
```

## Monitoring

### Metrics

Workers expose metrics:
- `worker_jobs_processed_total`: Total jobs processed
- `worker_jobs_failed_total`: Total jobs failed
- `worker_jobs_retry_total`: Total job retries
- `worker_jobs_dead_letter_total`: Total dead-letter jobs
- `worker_processing_duration_seconds`: Job processing duration
- `worker_poll_duration_seconds`: Poll duration

### Health Checks

Worker health endpoint:
```
GET /worker/healthz
```

Response:
```json
{
  "status": "healthy",
  "last_poll_at": "2024-01-01T12:00:00Z",
  "jobs_processed": 1000,
  "jobs_failed": 5,
  "dead_letter_count": 2
}
```

### Logging

Structured logging for:
- Job processing start/end
- Success/failure outcomes
- Retry attempts
- Dead-letter events
- Worker lifecycle events

Example:
```json
{
  "level": "info",
  "worker_type": "outbox_processor",
  "job_id": "uuid",
  "job_type": "order_created",
  "duration_ms": 150,
  "status": "success"
}
```

## Summary

The worker system provides:
- Reliable async processing without message broker
- Safe concurrent processing with SKIP LOCKED
- Retry logic with exponential backoff
- Dead-letter queue for failed jobs
- Idempotent processing
- Graceful shutdown
- Worker scaling support
- Comprehensive monitoring and logging
- Scheduled order activation with timezone validation
- Outbox pattern for reliable side effects
