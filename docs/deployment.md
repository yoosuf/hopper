# Deployment

## Overview

This document describes how to deploy the Food Delivery API Backend for local development, staging, and production environments.

## Local Development

### Prerequisites

- Go 1.25+
- PostgreSQL 16+
- Docker (optional, for containerized development)
- Make

### Setup

1. **Clone repository**:
```bash
git clone https://github.com/yourorg/hopper.git
cd hopper
```

2. **Copy environment variables**:
```bash
cp .env.example .env
```

3. **Edit .env file**:
```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=hopper
DB_PASSWORD=hopper_password
DB_NAME=hopper_db

# JWT
JWT_SECRET=your-secret-key-here-change-in-production

# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=60
```

4. **Start PostgreSQL**:
```bash
# Using Docker
docker-compose up -d postgres

# Or use local PostgreSQL
# Ensure PostgreSQL is running and create database
createdb hopper_db
```

5. **Run migrations**:
```bash
make migrate-up
```

6. **Seed database**:
```bash
make seed
```

7. **Run API server**:
```bash
make run
```

8. **Run worker** (in separate terminal):
```bash
make worker
```

### Docker Development

Using Docker Compose for local development:

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Rebuild after changes
docker-compose up -d --build
```

### Development Workflow

1. Make code changes
2. Run tests: `make test`
3. Run linting: `make lint`
4. Run migrations if needed: `make migrate-up`
5. Restart API/worker: `make run` / `make worker`

## Environment Variables

### Required Variables

```bash
# Database
DB_HOST=database-host
DB_PORT=5432
DB_USER=hopper
DB_PASSWORD=secure-password
DB_NAME=hopper_db
DB_SSL_MODE=require

# JWT
JWT_SECRET=your-secure-secret-key-min-32-chars

# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
ENVIRONMENT=development|staging|production

# CORS
CORS_ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com
CORS_ALLOWED_METHODS=GET,POST,PATCH,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization,Idempotency-Key
CORS_MAX_AGE=3600
```

### Optional Variables

```bash
# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=60
RATE_LIMIT_BURST=10
RATE_LIMIT_BY_IP=true
RATE_LIMIT_BY_USER=true

# Worker
WORKER_POLL_INTERVAL=5s
WORKER_BATCH_SIZE=100
WORKER_MAX_RETRIES=5
WORKER_RETRY_BASE_DELAY=1s
WORKER_CONCURRENCY=10

# Logging
LOG_LEVEL=info|debug|warn|error
LOG_FORMAT=json|text

# Metrics
METRICS_ENABLED=true
METRICS_PORT=9090

# SMTP (for notifications)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=notifications@example.com
SMTP_PASSWORD=smtp-password
SMTP_FROM=noreply@example.com

# Payment Provider
PAYMENT_PROVIDER=mock|stripe
PAYMENT_PROVIDER_SECRET=provider-secret-key
STRIPE_API_KEY=sk_test_...
```

## Migration Execution

### Using Migrate CLI

Install golang-migrate:
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Run migrations:
```bash
# Up
migrate -path migrations -database "postgres://user:password@localhost:5432/dbname?sslmode=disable" up

# Down
migrate -path migrations -database "postgres://user:password@localhost:5432/dbname?sslmode=disable" down 1

# Create new migration
migrate create -ext sql -dir migrations -seq migration_name
```

### Using Make

```bash
# Run all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Create new migration
make migrate-create name=add_new_table
```

### Migration Best Practices

- Always test migrations in development first
- Never modify existing migrations once deployed
- Use additive changes when possible
- Include down migration for rollback
- Test rollback in development
- Document breaking changes

## Worker Execution

### Running Worker

```bash
# Using Make
make worker

# Direct execution
go run cmd/worker/main.go

# Built binary
./bin/worker
```

### Worker Configuration

Workers can be enabled/disabled via environment variables:

```bash
# Enable all workers
WORKER_ENABLED=true

# Enable specific workers
OUTBOX_WORKER_ENABLED=true
SCHEDULED_ACTIVATOR_WORKER_ENABLED=true
NOTIFICATION_WORKER_ENABLED=true
PAYMENT_RECONCILER_WORKER_ENABLED=true
MAINTENANCE_WORKER_ENABLED=true
```

### Worker Scaling

For production, run multiple worker instances:

```bash
# Using systemd (example)
systemctl start hopper-worker@1
systemctl start hopper-worker@2
systemctl start hopper-worker@3

# Using Docker
docker-compose up --scale worker=3
```

### Worker Monitoring

Check worker health:
```bash
curl http://localhost:8081/healthz
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

## Regional Config Strategy

### Multi-Region Configuration

Each region has its own configuration in the database:

1. **Region record** (regions table):
   - timezone
   - currency_code
   - country_code
   - is_active

2. **Region config** (region_configs table):
   - platform_fee_basis_points
   - default_delivery_window_minutes
   - order_activation_lead_minutes
   - allow_scheduled_orders
   - delivery_fee_taxable_default

### Configuration via Database

All region-specific configuration stored in database, not code:

```sql
-- Example: Create region
INSERT INTO regions (code, name, country_code, timezone, currency_code, is_active)
VALUES ('US-NY', 'New York', 'US', 'America/New_York', 'USD', true);

-- Example: Configure region
INSERT INTO region_configs (region_id, platform_fee_basis_points, default_delivery_window_minutes, allow_scheduled_orders)
VALUES (region_uuid, 1500, 30, true);
```

### Configuration via Admin API

Region configuration managed via admin API:

```bash
# Create region
POST /v1/admin/regions
{
  "code": "US-NY",
  "name": "New York",
  "country_code": "US",
  "timezone": "America/New_York",
  "currency_code": "USD"
}

# Update region config
PATCH /v1/admin/region-configs/{id}
{
  "platform_fee_basis_points": 1500,
  "default_delivery_window_minutes": 30
}
```

### Configuration Caching

Region configuration cached in memory:
- Cache TTL: 1 hour
- Cache invalidated on configuration changes
- Fallback to database if cache miss

## Future Multi-Region Deployment Notes

### Single Region Deployment (Current)

Single database instance serving all regions:
- Simple to deploy
- Low operational overhead
- Suitable for initial scale

### Multi-Region Deployment (Future)

Deploy separate database per region:

**Architecture**:
```
Region US-East:
  - API Server (us-east-1)
  - PostgreSQL (us-east-1)
  - Workers (us-east-1)

Region EU-West:
  - API Server (eu-west-1)
  - PostgreSQL (eu-west-1)
  - Workers (eu-west-1)

Global:
  - Admin API (central)
  - Analytics (central)
  - Cross-region replication (for analytics)
```

**Data Partitioning**:
- Orders partitioned by region_id
- Users partitioned by primary region
- Restaurants partitioned by region_id

**Routing**:
- API Gateway routes requests to regional API servers
- User requests routed to user's primary region
- Cross-region requests via admin API

**Consistency**:
- Strong consistency within region
- Eventual consistency across regions
- Outbox pattern for cross-region events

### Migration Path

1. **Phase 1**: Single region, single database
2. **Phase 2**: Multi-region, single database (data isolation by region_id)
3. **Phase 3**: Multi-region, read replicas per region
4. **Phase 4**: Multi-region, separate databases per region

## Production Deployment Notes

### Deployment Strategy

**Blue-Green Deployment**:
- Two identical environments (blue, green)
- Deploy to green environment
- Test green environment
- Switch traffic from blue to green
- Keep blue as rollback target

**Rolling Updates**:
- Deploy to one instance at a time
- Health checks before routing traffic
- Gradual rollout with monitoring
- Automatic rollback on failure

### Infrastructure

**AWS Example**:
```
VPC:
  - Private subnets for API servers
  - Private subnets for database
  - Public subnets for load balancer

Database:
  - RDS PostgreSQL 16
  - Multi-AZ deployment
  - Automated backups
  - Read replicas for scaling

API Servers:
  - EC2 or ECS
  - Auto-scaling group
  - Load balancer (ALB)
  - SSL termination at ALB

Workers:
  - ECS or EC2
  - Auto-scaling group
  - SQS for job queue (future)

Monitoring:
  - CloudWatch metrics
  - CloudWatch logs
  - X-Ray tracing (optional)
```

**Kubernetes Example**:
```
Deployments:
  - hopper-api (replicas: 3)
  - hopper-worker (replicas: 2)

Services:
  - hopper-api-service (ClusterIP)
  - hopper-api-ingress (Ingress)

ConfigMaps:
  - hopper-config (environment variables)

Secrets:
  - hopper-secrets (sensitive data)

Persistent Volumes:
  - Not needed (stateless)
```

### Database

**Production Database Setup**:
```sql
-- Create database user with limited privileges
CREATE USER hopper_app WITH PASSWORD 'secure-password';

-- Grant privileges
GRANT CONNECT ON DATABASE hopper_db TO hopper_app;
GRANT USAGE ON SCHEMA public TO hopper_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO hopper_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO hopper_app;

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

**Connection Pooling**:
- Use PgBouncer for connection pooling
- Pool size: 20-50 connections
- Timeout: 30 seconds
- Max client connections: 1000

**Backups**:
- Daily automated backups
- Point-in-time recovery (7 days)
- Weekly full backups to S3
- Backup encryption at rest

### SSL/TLS

**API Server**:
- SSL termination at load balancer
- TLS 1.2+ only
- Strong cipher suites
- HSTS enabled
- Certificate auto-renewal

**Database**:
- SSL/TLS for database connections
- `DB_SSL_MODE=require`
- Certificate validation

### Secrets Management

**AWS Secrets Manager**:
```bash
# Store secrets
aws secretsmanager create-secret \
  --name hopper/prod \
  --secret-string '{"JWT_SECRET":"...","DB_PASSWORD":"..."}'

# Retrieve secrets in application
```

**HashiCorp Vault**:
```bash
# Store secrets
vault kv put secret/hopper/prod JWT_SECRET=... DB_PASSWORD=...

# Retrieve secrets in application
```

### Monitoring

**Metrics**:
- Request rate, error rate, latency
- Database connection pool usage
- Worker job processing metrics
- System metrics (CPU, memory, disk)

**Logging**:
- Structured JSON logs
- Centralized log aggregation (ELK, CloudWatch)
- Log retention: 30 days
- Sensitive data redacted

**Alerting**:
- High error rate (> 1%)
- High latency (> 1s p95)
- Database connection exhaustion
- Worker queue backlog
- Dead-letter queue accumulation

### Health Checks

**API Health Check**:
```
GET /healthz
```

Response:
```json
{
  "status": "healthy",
  "database": "connected",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

**Readiness Check**:
```
GET /readyz
```

Response:
```json
{
  "status": "ready",
  "database": "ready",
  "workers": "ready"
}
```

### Graceful Shutdown

**Shutdown Flow**:
1. Stop accepting new requests
2. Drain in-flight requests (timeout: 30s)
3. Stop worker polling
4. Complete in-progress jobs
5. Close database connections
6. Exit cleanly

**Implementation**:
```go
func (s *Server) Shutdown(ctx context.Context) error {
    // Stop accepting new requests
    s.server.Shutdown(ctx)
    
    // Drain workers
    s.worker.Drain()
    
    // Close connections
    s.db.Close()
    
    return nil
}
```

### Scaling

**Horizontal Scaling**:
- API servers: Auto-scaling based on CPU/memory
- Workers: Auto-scaling based on queue depth
- Database: Read replicas for read-heavy workloads

**Vertical Scaling**:
- Increase instance size for API servers
- Increase database instance class
- Add more CPU/memory to workers

### Disaster Recovery

**Backup Strategy**:
- Daily automated backups
- Point-in-time recovery (7 days)
- Cross-region backup replication
- Regular restore testing

**Failover**:
- Multi-AZ database deployment
- Automatic failover to standby
- RTO: 1 hour
- RPO: 5 minutes

## Production Checklist

### Pre-Deployment

- [ ] All tests passing
- [ ] Security audit completed
- [ ] Dependencies scanned for vulnerabilities
- [ ] Environment variables configured
- [ ] Secrets stored in secret manager
- [ ] Database backups verified
- [ ] SSL certificates valid
- [ ] Monitoring and alerting configured
- [ ] Log aggregation configured
- [ ] Rate limiting configured
- [ ] CORS configured
- [ ] Health check endpoints working

### Post-Deployment

- [ ] Health checks passing
- [ ] Database connections healthy
- [ ] Workers processing jobs
- [ ] No errors in logs
- [ ] Metrics within normal range
- [ ] Smoke tests passing
- [ ] Performance baseline established

### Ongoing

- [ ] Daily backup verification
- [ ] Weekly dependency updates
- [ ] Monthly security audits
- [ ] Quarterly disaster recovery testing
- [ ] Regular log review
- [ ] Performance tuning

## Troubleshooting

### Database Connection Issues

**Symptoms**: API returns database connection errors

**Solutions**:
1. Check database is running: `docker-compose ps postgres`
2. Check connection string in .env
3. Check database credentials
4. Check network connectivity
5. Check connection pool limits

### Worker Not Processing Jobs

**Symptoms**: Jobs accumulating in pending status

**Solutions**:
1. Check worker is running: `ps aux | grep worker`
2. Check worker logs
3. Check database connection
4. Check worker configuration
5. Check job table for locked jobs

### Migration Failures

**Symptoms**: Migration fails to apply

**Solutions**:
1. Check migration file syntax
2. Check database permissions
3. Check for existing data conflicts
4. Rollback and investigate
5. Test migration in development first

### High Memory Usage

**Symptoms**: API or worker using excessive memory

**Solutions**:
1. Check for memory leaks
2. Check connection pool size
3. Check for goroutine leaks
4. Profile application
5. Increase memory limits

### Slow Performance

**Symptoms**: API response times high

**Solutions**:
1. Check database query performance
2. Check for missing indexes
3. Check connection pool exhaustion
4. Check for N+1 queries
5. Add caching where appropriate

## Summary

Deployment includes:
- Local development with Docker Compose
- Environment variable configuration
- Database migration management
- Worker execution and scaling
- Regional configuration strategy
- Production deployment notes
- Multi-region future considerations
- Comprehensive monitoring and alerting
- Disaster recovery planning
- Troubleshooting guide
