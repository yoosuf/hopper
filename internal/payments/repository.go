package payments

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for payment data access
type Repository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*Payment, error)
	ListByOrder(ctx context.Context, orderID uuid.UUID) ([]*Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, providerID *string) error
}

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRepository creates a new payments repository
func NewRepository(db *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

// Create creates a new payment
func (r *RepositoryImpl) Create(ctx context.Context, payment *Payment) error {
	query := `
		INSERT INTO payments (id, order_id, amount, currency_code, status, payment_method, payment_provider_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW() AT TIME ZONE 'UTC', NOW() AT TIME ZONE 'UTC')
	`

	_, err := r.db.Exec(ctx, query,
		payment.ID,
		payment.OrderID,
		payment.Amount,
		payment.CurrencyCode,
		payment.Status,
		payment.PaymentMethod,
		payment.PaymentProviderID,
	)

	return err
}

// GetByID retrieves a payment by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*Payment, error) {
	query := `
		SELECT id, order_id, amount, currency_code, status, payment_method, payment_provider_id
		FROM payments
		WHERE id = $1 AND deleted_at IS NULL
	`

	var payment Payment
	err := r.db.QueryRow(ctx, query, id).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.Amount,
		&payment.CurrencyCode,
		&payment.Status,
		&payment.PaymentMethod,
		&payment.PaymentProviderID,
	)

	if err != nil {
		return nil, err
	}

	return &payment, nil
}

// ListByOrder lists payments for an order
func (r *RepositoryImpl) ListByOrder(ctx context.Context, orderID uuid.UUID) ([]*Payment, error) {
	query := `
		SELECT id, order_id, amount, currency_code, status, payment_method, payment_provider_id
		FROM payments
		WHERE order_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*Payment
	for rows.Next() {
		var payment Payment
		err := rows.Scan(
			&payment.ID,
			&payment.OrderID,
			&payment.Amount,
			&payment.CurrencyCode,
			&payment.Status,
			&payment.PaymentMethod,
			&payment.PaymentProviderID,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, &payment)
	}

	return payments, nil
}

// UpdateStatus updates a payment's status
func (r *RepositoryImpl) UpdateStatus(ctx context.Context, id uuid.UUID, status string, providerID *string) error {
	query := `
		UPDATE payments
		SET status = $2, payment_provider_id = $3, updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, status, providerID)
	return err
}
