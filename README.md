# Hopper - Food Delivery API Backend

A production-grade, multi-region food delivery API backend built with Go and PostgreSQL.

## Features

- **Multi-Region Support**: Support for restaurants and deliveries across different geographic regions
- **Tax-Aware Pricing**: Automatic tax calculation based on regional tax zones and categories
- **Role-Based Access Control**: Customer, restaurant owner, courier, and admin roles with scoped permissions
- **Workflow State Management**: Order and delivery lifecycle management with state transitions
- **Background Workers**: Asynchronous job processing for notifications and background tasks
- **Audit Logging**: Complete audit trail for all critical operations
- **RESTful API**: Clean, well-documented REST API with OpenAPI specification
- **Docker Support**: Containerized deployment with Docker and Docker Compose

## Technology Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL 15+
- **HTTP Router**: Chi
- **Database Driver**: pgx (pgxpool)
- **Authentication**: JWT (JSON Web Tokens)
- **Validation**: go-playground/validator
- **Metrics**: Prometheus-compatible metrics

## Project Structure

```
hopper/
├── cmd/
│   ├── api/          # API server entry point
│   └── worker/       # Background worker entry point
├── internal/
│   ├── admin/        # Admin operations
│   ├── audit/        # Audit logging
│   ├── auth/         # Authentication service
│   ├── delivery/     # Delivery management
│   ├── jobs/         # Job scheduling
│   ├── menus/        # Menu item management
│   ├── notifications/# Notification service
│   ├── orders/       # Order processing
│   ├── payments/     # Payment processing
│   ├── regions/      # Regional configuration
│   ├── restaurants/  # Restaurant management
│   ├── tax/          # Tax calculation
│   ├── users/        # User management
│   ├── worker/       # Background worker
│   └── platform/     # Shared platform components
│       ├── config/   # Configuration
│       ├── db/       # Database connection
│       ├── httpx/    # HTTP utilities
│       ├── logger/   # Logging
│       ├── metrics/  # Metrics collection
│       ├── middleware/# HTTP middleware
│       ├── validator/# Request validation
│       ├── idempotency/# Idempotency
│       └── clock/    # Time utilities
├── docs/             # Documentation
├── openapi/          # OpenAPI specification
├── scripts/          # Utility scripts
├── migrations/       # Database migrations
├── Dockerfile        # Docker image definition
├── docker-compose.yml # Docker Compose configuration
├── Makefile          # Build automation
└── go.mod            # Go module definition
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 15 or higher
- Docker and Docker Compose (for containerized deployment)

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/crewdigital/hopper.git
   cd hopper
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Configure environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run database migrations**
   ```bash
   make migrate
   ```

5. **Seed database (optional)**
   ```bash
   make seed
   ```

6. **Run the API server**
   ```bash
   make run-api
   ```

7. **Run the worker (in separate terminal)**
   ```bash
   make run-worker
   ```

### Docker Deployment

1. **Build and start with Docker Compose**
   ```bash
   make docker-up
   ```

2. **View logs**
   ```bash
   make docker-logs
   ```

3. **Stop services**
   ```bash
   make docker-down
   ```

## API Endpoints

### Authentication
- `POST /auth/register` - Register a new user
- `POST /auth/login` - Login and get JWT token
- `POST /auth/refresh` - Refresh JWT token

### Users
- `GET /users/me` - Get current user
- `PUT /users/me` - Update current user

### Regions
- `GET /regions` - List all regions
- `GET /regions/{id}` - Get region details

### Tax
- `GET /tax/categories` - List tax categories
- `GET /tax/zones` - List tax zones
- `GET /tax/rates` - List tax rates

### Restaurants
- `GET /restaurants` - List restaurants
- `POST /restaurants` - Create restaurant
- `GET /restaurants/{id}` - Get restaurant details
- `PUT /restaurants/{id}` - Update restaurant
- `DELETE /restaurants/{id}` - Delete restaurant

### Menus
- `GET /restaurants/{restaurant_id}/menus` - List menu items
- `POST /restaurants/{restaurant_id}/menus` - Create menu item
- `PUT /restaurants/{restaurant_id}/menus/{id}` - Update menu item
- `DELETE /restaurants/{restaurant_id}/menus/{id}` - Delete menu item

### Orders
- `POST /orders` - Create order
- `GET /orders/{id}` - Get order details
- `GET /orders` - List customer orders
- `POST /orders/{id}/cancel` - Cancel order

### Deliveries
- `GET /deliveries/{id}` - Get delivery details
- `GET /couriers/deliveries` - List courier deliveries
- `PUT /deliveries/{id}/status` - Update delivery status

### Payments
- `POST /payments` - Create payment
- `GET /payments/{id}` - Get payment details
- `PUT /payments/{id}/status` - Update payment status

### Notifications
- `GET /notifications` - List user notifications
- `PUT /notifications/{id}/read` - Mark notification as read

### Admin
- `POST /admin/restaurants/{id}/approve` - Approve restaurant
- `POST /admin/restaurants/{id}/reject` - Reject restaurant
- `GET /admin/stats` - Get system statistics

For detailed API documentation, see [docs/api.md](docs/api.md) or [openapi/openapi.yaml](openapi/openapi.yaml).

## Configuration

The application is configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | - |
| `JWT_SECRET` | JWT signing secret | - |
| `PORT` | API server port | `8080` |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` |

## Testing

Run all tests:
```bash
make test
```

Run tests with coverage:
```bash
make test-coverage
```

## Build

Build the application binaries:
```bash
make build
```

## Code Quality

Format code:
```bash
make fmt
```

Run linter:
```bash
make lint
```

Run all checks:
```bash
make check
```

## Documentation

- [Architecture](docs/architecture.md) - System architecture overview
- [Design](docs/design.md) - Detailed design decisions
- [Database](docs/database.md) - Database schema and design
- [API](docs/api.md) - API documentation
- [Workers](docs/workers.md) - Background worker documentation
- [Security](docs/security.md) - Security considerations
- [Deployment](docs/deployment.md) - Deployment guide
- [Testing](docs/testing.md) - Testing strategy
- [Tax](docs/tax.md) - Tax calculation details

## License

Copyright (c) Crew Digital. All rights reserved.
