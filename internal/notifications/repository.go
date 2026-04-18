package notifications

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for notification data access
type Repository interface {
	Create(ctx context.Context, notification *Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Notification, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	GetPreference(ctx context.Context, userID uuid.UUID, channel NotificationChannel) (*NotificationPreference, error)
	SetPreference(ctx context.Context, pref *NotificationPreference) error
	ListPreferences(ctx context.Context, userID uuid.UUID) ([]*NotificationPreference, error)
}

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRepository creates a new notifications repository
func NewRepository(db *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

// Create creates a new notification
func (r *RepositoryImpl) Create(ctx context.Context, notification *Notification) error {
	query := `
		INSERT INTO notifications (id, user_id, type, title, body, data, is_read, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW() AT TIME ZONE 'UTC', NOW() AT TIME ZONE 'UTC')
	`

	_, err := r.db.Exec(ctx, query,
		notification.ID,
		notification.UserID,
		notification.Type,
		notification.Title,
		notification.Body,
		notification.Data,
		notification.IsRead,
	)

	return err
}

// GetByID retrieves a notification by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*Notification, error) {
	query := `
		SELECT id, user_id, type, title, body, data, is_read, created_at
		FROM notifications
		WHERE id = $1 AND deleted_at IS NULL
	`

	var notification Notification
	err := r.db.QueryRow(ctx, query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Title,
		&notification.Body,
		&notification.Data,
		&notification.IsRead,
		&notification.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &notification, nil
}

// ListByUser lists notifications for a user
func (r *RepositoryImpl) ListByUser(ctx context.Context, userID uuid.UUID) ([]*Notification, error) {
	query := `
		SELECT id, user_id, type, title, body, data, is_read, created_at
		FROM notifications
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		var notification Notification
		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Body,
			&notification.Data,
			&notification.IsRead,
			&notification.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read
func (r *RepositoryImpl) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE notifications
		SET is_read = true, updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	return err
}

// MarkAllAsRead marks all notifications for a user as read
func (r *RepositoryImpl) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET is_read = true, updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE user_id = $1 AND is_read = false
	`

	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// GetPreference retrieves a notification preference
func (r *RepositoryImpl) GetPreference(ctx context.Context, userID uuid.UUID, channel NotificationChannel) (*NotificationPreference, error) {
	query := `
		SELECT user_id, channel, enabled, order_events, payment_events
		FROM notification_preferences
		WHERE user_id = $1 AND channel = $2
	`

	var pref NotificationPreference
	err := r.db.QueryRow(ctx, query, userID, channel).Scan(
		&pref.UserID,
		&pref.Channel,
		&pref.Enabled,
		&pref.OrderEvents,
		&pref.PaymentEvents,
	)
	if err != nil {
		return nil, err
	}

	return &pref, nil
}

// SetPreference sets a notification preference
func (r *RepositoryImpl) SetPreference(ctx context.Context, pref *NotificationPreference) error {
	query := `
		INSERT INTO notification_preferences (user_id, channel, enabled, order_events, payment_events, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW() AT TIME ZONE 'UTC', NOW() AT TIME ZONE 'UTC')
		ON CONFLICT (user_id, channel)
		DO UPDATE SET enabled = $3, order_events = $4, payment_events = $5, updated_at = NOW() AT TIME ZONE 'UTC'
	`

	_, err := r.db.Exec(ctx, query,
		pref.UserID,
		pref.Channel,
		pref.Enabled,
		pref.OrderEvents,
		pref.PaymentEvents,
	)
	return err
}

// ListPreferences retrieves all notification preferences for a user
func (r *RepositoryImpl) ListPreferences(ctx context.Context, userID uuid.UUID) ([]*NotificationPreference, error) {
	query := `
		SELECT user_id, channel, enabled, order_events, payment_events
		FROM notification_preferences
		WHERE user_id = $1
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefs []*NotificationPreference
	for rows.Next() {
		var pref NotificationPreference
		err := rows.Scan(
			&pref.UserID,
			&pref.Channel,
			&pref.Enabled,
			&pref.OrderEvents,
			&pref.PaymentEvents,
		)
		if err != nil {
			return nil, err
		}
		prefs = append(prefs, &pref)
	}

	return prefs, nil
}
