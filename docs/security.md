# Security

## Overview

This document describes the security posture, authentication, authorization, data protection, and compliance measures for the Food Delivery API Backend.

## Recent Security Enhancements

### CSRF Protection
- CSRF middleware enabled when Redis is configured
- Token-based CSRF validation for state-changing operations
- Origin validation to prevent cross-site request forgery attacks
- Configurable via `REDIS_ENABLED` environment variable

### Distributed Rate Limiting
- Cache-based rate limiting supporting both in-memory and Redis backends
- Configurable request limits per minute per IP
- Automatic fallback to in-memory rate limiting when Redis is unavailable
- Distributed rate limiting enabled with Redis for multi-instance deployments

### AppError Pattern
- Consistent error handling across all services
- Type-safe error codes and messages
- Prevents fragile error string comparisons
- Improved error tracking and debugging

### Structured Logging
- JSON-based structured logging for security event tracking
- Configurable log levels (debug, info, warn, error)
- Request ID correlation for audit trails
- Sensitive data filtering in logs

## Password Hashing

### Algorithm

We use bcrypt for password hashing with a cost factor of 12.

**Rationale**:
- Proven, battle-tested algorithm
- Built-in salt generation
- Adjustable cost factor for performance/security balance
- Resistant to brute force and rainbow table attacks

### Implementation

```go
import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    return string(bytes), err
}

func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

### Password Requirements

- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number
- At least one special character

### Storage

- Only password hashes stored in database
- Raw passwords never stored
- Password hashes never logged
- Password hashes never exposed in API responses

## JWT Strategy

### Access Tokens

**Purpose**: Short-lived authentication tokens for API access

**Configuration**:
- Algorithm: RS256 (RSA with SHA-256) or HS256 (HMAC with SHA-256)
- TTL: 15 minutes
- Issuer: API service identifier
- Audience: API service identifier

**Claims**:
```json
{
  "sub": "user_id",
  "role": "customer|restaurant_owner|courier|admin",
  "iat": 1234567890,
  "exp": 1234567890,
  "iss": "hopper-api",
  "aud": "hopper-api"
}
```

**Signing**:
- Private key stored securely (environment variable or secret manager)
- Public key used for verification (if RS256)
- Secret key used for signing/verification (if HS256)

### Refresh Tokens

**Purpose**: Long-lived tokens for obtaining new access tokens

**Configuration**:
- TTL: 7 days (configurable)
- Stored in database (`refresh_tokens` table)
- Token hash stored (SHA-256)
- Single use (rotated on each refresh)

**Storage Schema**:
```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    revoked_at TIMESTAMP WITH TIME ZONE
);
```

### Refresh Token Rotation

**Flow**:
1. User presents refresh token
2. Server validates token (not revoked, not expired)
3. Server generates new access token
4. Server generates new refresh token
5. Server revokes old refresh token
6. Server returns new access + refresh tokens

**Benefits**:
- Limits window of token compromise
- Detects token theft (old token reuse)
- Automatic token lifecycle management

### Refresh Token Revocation

**Logout Flow**:
1. User presents refresh token
2. Server marks token as revoked (`revoked_at = now()`)
3. Server returns success
4. Token cannot be used again

**Admin Revocation**:
- Admin can revoke all refresh tokens for a user
- Used for security incidents or account compromise

## Role/Ownership Enforcement

### Role-Based Access Control (RBAC)

**Middleware**:
```go
func RequireRole(roles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userRole := getUserRoleFromContext(r.Context())
            
            if !contains(roles, userRole) {
                respondError(w, http.StatusForbidden, "insufficient permissions")
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

**Usage**:
```go
router.With(authMiddleware, RequireRole("customer")).HandleFunc("/v1/me/orders", handler)
router.With(authMiddleware, RequireRole("restaurant_owner")).HandleFunc("/v1/owner/orders", handler)
router.With(authMiddleware, RequireRole("admin")).HandleFunc("/v1/admin/users", handler)
```

### Ownership Checks

**Service Layer**:
```go
func (s *OrderService) GetOrder(ctx context.Context, userID, orderID uuid.UUID) (*Order, error) {
    order, err := s.repo.GetByID(ctx, orderID)
    if err != nil {
        return nil, err
    }
    
    // Ownership check
    if order.CustomerID != userID {
        return nil, ErrForbidden
    }
    
    return order, nil
}
```

**Repository Layer**:
```go
func (r *OrderRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID, orderID uuid.UUID) (*Order, error) {
    query := `
        SELECT * FROM orders
        WHERE id = $1 AND customer_id = $2
    `
    return r.queryOne(ctx, query, orderID, customerID)
}
```

### Object-Level Security

**Prevention of IDOR**:
- Always validate ownership in service layer
- Use scoped queries in repository layer
- Never expose internal IDs in URLs without validation
- Use UUIDs instead of sequential integers

**Example**:
```go
// BAD: Vulnerable to IDOR
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
    orderID := getIDFromURL(r)
    order, _ := h.repo.GetByID(r.Context(), orderID) // No ownership check!
    respondJSON(w, order)
}

// GOOD: Secure
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
    userID := getUserIDFromContext(r.Context())
    orderID := getIDFromURL(r)
    order, err := h.service.GetOrder(r.Context(), userID, orderID) // Ownership check in service
    if err != nil {
        respondError(w, http.StatusForbidden, "access denied")
        return
    }
    respondJSON(w, order)
}
```

## Validation Approach

### Request Validation

**Layer**: Handler layer before business logic

**Library**: go-playground/validator

**Validation Rules**:
- Email format validation
- Password complexity validation
- Required field validation
- Field length validation
- Enum value validation
- Numeric range validation

**Example**:
```go
type CreateOrderRequest struct {
    RestaurantID uuid.UUID `json:"restaurant_id" validate:"required"`
    DeliveryType string    `json:"delivery_type" validate:"required,oneof=asap scheduled"`
    Items       []OrderItemRequest `json:"items" validate:"required,min=1"`
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    var req CreateOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "invalid request body")
        return
    }
    
    if err := h.validator.Struct(&req); err != nil {
        respondValidationError(w, err)
        return
    }
    
    // Process request
}
```

### DTO Validation

**Strict DTOs**: Never bind request directly to DB models

**Example**:
```go
// BAD: Direct binding to DB model
type User struct {
    ID           uuid.UUID `json:"id"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"password_hash"` // Exposed!
    Role         string    `json:"role"`
}

// GOOD: Separate DTOs
type UserResponse struct {
    ID    uuid.UUID `json:"id"`
    Email string    `json:"email"`
    Role  string    `json:"role"`
    // PasswordHash not exposed
}
```

### Input Sanitization

**Sanitization Rules**:
- Trim whitespace from string inputs
- Normalize email addresses (lowercase)
- Sanitize HTML in free-text fields (prevent XSS)
- Validate and sanitize file uploads (not in MVP)

**Example**:
```go
func sanitizeInput(input string) string {
    input = strings.TrimSpace(input)
    input = html.EscapeString(input)
    return input
}
```

## Rate Limiting

### Rate Limiting Strategy

**Algorithm**: Token bucket or sliding window

**Implementation**: In-memory with Redis (future) or in-memory map (MVP)

**Configuration**:
```
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=60
RATE_LIMIT_BURST=10
RATE_LIMIT_BY_IP=true
RATE_LIMIT_BY_USER=true
```

### Rate Limiting by IP

**Middleware**:
```go
func RateLimitByIP(limiter *RateLimiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ip := getClientIP(r)
            
            if !limiter.Allow(ip) {
                respondError(w, http.StatusTooManyRequests, "rate limit exceeded")
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Rate Limiting by User

**Middleware**:
```go
func RateLimitByUser(limiter *RateLimiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := getUserIDFromContext(r.Context())
            
            if !limiter.Allow(userID.String()) {
                respondError(w, http.StatusTooManyRequests, "rate limit exceeded")
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Rate Limiting Endpoints

**High-Risk Endpoints** (stricter limits):
- POST /v1/auth/login: 5 requests per minute
- POST /v1/auth/register: 3 requests per minute
- POST /v1/orders: 20 requests per minute
- POST /v1/courier/deliveries/{id}/claim: 10 requests per minute

**Standard Endpoints** (normal limits):
- GET /v1/restaurants: 60 requests per minute
- GET /v1/me/orders: 60 requests per minute

### Response Headers

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1234567890
```

## Idempotency

### Idempotency Key Support

**Purpose**: Prevent duplicate processing of critical operations

**Supported Endpoints**:
- POST /v1/orders
- POST /v1/payments (initiate)

### Implementation

**Middleware**:
```go
func Idempotency(service *IdempotencyService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            key := r.Header.Get("Idempotency-Key")
            if key == "" {
                next.ServeHTTP(w, r)
                return
            }
            
            scope := getUserIDFromContext(r.Context()).String()
            
            // Check if key already used
            record, err := service.Get(r.Context(), key, scope)
            if err == nil {
                // Return cached response
                w.Header().Set("X-Idempotency-Key", key)
                w.WriteHeader(record.ResponseCode)
                w.Write(record.ResponseBody)
                return
            }
            
            // Capture response
            recorder := httptest.NewRecorder()
            next.ServeHTTP(recorder, r)
            
            // Store response
            service.Store(r.Context(), key, scope, r, recorder)
            
            // Return response
            w.Header().Set("X-Idempotency-Key", key)
            for k, v := range recorder.Header() {
                w.Header()[k] = v
            }
            w.WriteHeader(recorder.Code)
            w.Write(recorder.Body.Bytes())
        })
    }
}
```

### Storage Schema

```sql
CREATE TABLE idempotency_keys (
    id UUID PRIMARY KEY,
    key VARCHAR(255) NOT NULL,
    scope VARCHAR(100) NOT NULL,
    request_hash VARCHAR(64),
    response_code INTEGER,
    response_body JSONB,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    UNIQUE (key, scope)
);
```

### Expiration

- Idempotency records expire after 24 hours
- Cleanup job removes expired records
- Configurable per environment

## Audit Logging

### Audit Log Schema

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    actor_id UUID NOT NULL,
    actor_role VARCHAR(50) NOT NULL,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID,
    metadata JSONB,
    ip_address INET,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);
```

### Audited Actions

**Admin Actions**:
- User creation/suspension/restoration
- Restaurant approval/suspension/archival
- Region configuration changes
- Tax configuration changes
- Order status overrides
- Payment status overrides

**Business Actions**:
- Restaurant approval
- Role changes
- Order cancellation (by admin)
- Payment status changes

### Audit Log Creation

**Service Layer**:
```go
func (s *AdminService) ApproveRestaurant(ctx context.Context, restaurantID uuid.UUID, actorID uuid.UUID) error {
    // Perform approval
    if err := s.restaurantRepo.Approve(ctx, restaurantID); err != nil {
        return err
    }
    
    // Create audit log
    auditLog := &AuditLog{
        ActorID:    actorID,
        ActorRole:  getRoleFromContext(ctx),
        Action:     "restaurant_approved",
        EntityType: "restaurant",
        EntityID:   restaurantID,
        IPAddress:  getIPFromContext(ctx),
    }
    
    return s.auditRepo.Create(ctx, auditLog)
}
```

### Audit Log Query

**Admin Endpoint**:
```
GET /v1/admin/audit-logs?entity_type=restaurant&entity_id={id}
```

**Filters**:
- actor_id
- actor_role
- action
- entity_type
- entity_id
- date range

## Secrets Management

### Environment Variables

**Required Secrets**:
```
JWT_SECRET=your-secret-key-here
DB_PASSWORD=your-database-password
SMTP_PASSWORD=your-smtp-password
PAYMENT_PROVIDER_SECRET=your-payment-secret
```

### Secret Rotation Strategy

**JWT Secret**:
- Rotate every 90 days
- Support multiple active secrets during rotation
- Gradual rollout of new secret
- Old secret valid for 7 days after rotation

**Database Password**:
- Rotate every 90 days
- Requires application restart
- Coordinate with DBA team

**API Keys**:
- Rotate on compromise suspicion
- Rotate every 180 days
- Support key versioning

### Secret Storage

**Development**:
- Environment variables in `.env` file
- `.env` in `.gitignore`

**Production**:
- Use secret manager (AWS Secrets Manager, HashiCorp Vault)
- Inject secrets at runtime
- Never commit secrets to version control

## CORS and Headers

### CORS Configuration

**Environment Variables**:
```
CORS_ALLOWED_ORIGINS=https://example.com,https://app.example.com
CORS_ALLOWED_METHODS=GET,POST,PATCH,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization,Idempotency-Key
CORS_MAX_AGE=3600
```

**Middleware**:
```go
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            
            if isAllowedOrigin(origin, allowedOrigins) {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,Idempotency-Key")
                w.Header().Set("Access-Control-Max-Age", "3600")
            }
            
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusNoContent)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Security Headers

**Middleware**:
```go
func SecurityHeaders() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("X-Content-Type-Options", "nosniff")
            w.Header().Set("X-Frame-Options", "DENY")
            w.Header().Set("X-XSS-Protection", "1; mode=block")
            w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
            w.Header().Set("Content-Security-Policy", "default-src 'self'")
            w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
            
            next.ServeHTTP(w, r)
        })
    }
}
```

## Production Hardening Checklist

### TLS/SSL

- [ ] TLS 1.2+ only (disable TLS 1.0, 1.1)
- [ ] Strong cipher suites
- [ ] Valid certificates from trusted CA
- [ ] Certificate auto-renewal
- [ ] HSTS enabled

### Database Security

- [ ] Strong password (32+ characters, mixed case, numbers, symbols)
- [ ] SSL/TLS for database connections
- [ ] Restricted database user (least privilege)
- [ ] No direct database access from internet
- [ ] Database backups encrypted
- [ ] Connection pooling limits

### Application Security

- [ ] Dependency scanning (Snyk, Dependabot)
- [ ] Static code analysis (SonarQube, Gosec)
- [ ] Regular security audits
- [ ] Penetration testing
- [ ] Bug bounty program
- [ ] Security incident response plan

### Infrastructure Security

- [ ] VPC/private network isolation
- [ ] Firewall rules (allow only necessary ports)
- [ ] Security groups (restrict access)
- [ ] IAM roles (least privilege)
- [ ] MFA for admin access
- [ ] Audit logging for infrastructure changes
- [ ] Immutable infrastructure

### Monitoring and Alerting

- [ ] Security event logging
- [ ] Anomaly detection
- [ ] Failed login alerts
- [ ] Rate limit breach alerts
- [ ] Unusual API usage alerts
- [ ] Database connection alerts
- [ ] 24/7 security monitoring

## PII Handling

### Data Classification

**Highly Sensitive**:
- Password hashes (never exposed)
- Refresh tokens (never exposed)
- Payment provider tokens (never exposed)

**Sensitive**:
- Email addresses
- Phone numbers
- Delivery addresses
- Payment information (non-card)

**Public**:
- Restaurant names
- Menu items
- Order totals (non-PII)

### PII Minimization

- Collect only necessary PII
- Anonymize logs (redact PII)
- Data retention policies
- Right to deletion (GDPR)
- Data export (GDPR)

### Data Masking

**Log Redaction**:
```go
func redactPII(log string) string {
    log = redactEmail(log)
    log = redactPhone(log)
    log = redactAddress(log)
    log = redactToken(log)
    return log
}
```

**Response DTOs**:
- Never expose internal IDs in error messages
- Never expose stack traces in production
- Never expose PII in list responses (only detail views)

## Payment Security Boundaries

### PCI DSS Scope

**Out of Scope**:
- Raw card data never stored
- Card data never logged
- Card data never transmitted through our servers

**In Scope**:
- Payment provider tokenization
- Payment status tracking
- Payment reconciliation

### Payment Provider Abstraction

**Interface**:
```go
type PaymentProvider interface {
    CreatePaymentToken(ctx context.Context, cardData CardData) (string, error)
    ChargePayment(ctx context.Context, token string, amount int, currency string) (PaymentResult, error)
    RefundPayment(ctx context.Context, paymentID string) (RefundResult, error)
    GetPaymentStatus(ctx context.Context, paymentID string) (PaymentStatus, error)
}
```

**Mock Provider** (MVP):
- Simulates payment processing
- Returns success/failure based on configuration
- No real card data handled

**Stripe Integration** (Production):
- Uses Stripe API
- Stripe handles card data
- We only store Stripe payment method tokens

### Payment Flow

1. Client sends card data to Stripe (direct)
2. Stripe returns payment method token
3. Client sends token to our API
4. Our API creates payment with token
5. Our API never sees raw card data

## Region/Compliance Notes

### GDPR (EU)

- Data processing agreements
- Right to access
- Right to deletion
- Right to portability
- Data breach notification (72 hours)
- Privacy policy
- Cookie consent

### CCPA (California)

- Right to know
- Right to delete
- Right to opt-out
- Right to non-discrimination
- Privacy policy
- Do not sell my info

### Multi-Region Data Residency

- Data stored in region-specific databases (future)
- Cross-border data transfer compliance
- Regional data retention policies
- Regional audit requirements

## Audit Implications for Tax/Admin Changes

### Tax Configuration Changes

**Audited Actions**:
- Tax rate creation/update/deletion
- Tax zone creation/update/deletion
- Tax category creation/update/deletion
- Region configuration changes

**Audit Requirements**:
- Actor ID and role
- Previous value (if update)
- New value
- Reason for change
- Timestamp

**Example**:
```json
{
  "actor_id": "admin_uuid",
  "actor_role": "admin",
  "action": "tax_rate_updated",
  "entity_type": "tax_rate",
  "entity_id": "tax_rate_uuid",
  "metadata": {
    "previous_rate_bps": 800,
    "new_rate_bps": 825,
    "reason": "State tax rate increase"
  },
  "created_at": "2024-01-01T12:00:00Z"
}
```

### Admin Overrides

**Audited Actions**:
- Order status override
- Payment status override
- User suspension/activation
- Restaurant approval/suspension

**Audit Requirements**:
- Actor ID and role
- Override reason
- Previous state
- New state
- Affected entities

## Broken Object-Level Authorization Prevention

### Prevention Strategies

1. **Always validate ownership in service layer**
2. **Use scoped queries in repository layer**
3. **Never trust client-provided IDs**
4. **Use UUIDs instead of sequential integers**
5. **Implement RBAC middleware**
6. **Regular security audits**
7. **Automated testing for IDOR vulnerabilities**

### Testing for IDOR

**Test Cases**:
- User A tries to access User B's orders
- Restaurant Owner A tries to access Restaurant Owner B's menu
- Courier A tries to claim Courier B's delivery
- Customer tries to access admin endpoints

**Example Test**:
```go
func TestOrderOwnership(t *testing.T) {
    userA := createTestUser(t, "customer")
    userB := createTestUser(t, "customer")
    order := createTestOrder(t, userA.ID)
    
    // User B tries to access User A's order
    req := httptest.NewRequest("GET", "/v1/orders/"+order.ID.String(), nil)
    req = req.WithContext(contextWithUser(req.Context(), userB))
    
    resp := executeRequest(req)
    
    assert.Equal(t, http.StatusForbidden, resp.Code)
}
```

## Summary

The security posture includes:
- Strong password hashing (bcrypt)
- JWT access tokens with short TTL
- Refresh token rotation and revocation
- Role-based access control (RBAC)
- Ownership checks on all CRUD operations
- Object-level security (IDOR prevention)
- Request validation with strict DTOs
- Rate limiting by IP and user
- Idempotency for critical operations
- Comprehensive audit logging
- Secrets management via environment variables
- CORS and security headers
- Payment security boundaries (PCI DSS out of scope)
- PII handling and minimization
- Multi-region compliance support
- Regular security audits and testing
