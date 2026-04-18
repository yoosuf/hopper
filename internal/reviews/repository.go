package reviews

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Review represents a review
type Review struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	TargetType   string // "restaurant", "courier", "menu_item"
	TargetID     uuid.UUID
	Rating       int    // 1-5 stars
	Comment      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Repository defines the interface for review data access
type Repository interface {
	Create(ctx context.Context, review *Review) error
	GetByID(ctx context.Context, id uuid.UUID) (*Review, error)
	GetByUserAndTarget(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) (*Review, error)
	ListByTarget(ctx context.Context, targetType string, targetID uuid.UUID, limit int) ([]*Review, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*Review, error)
	Update(ctx context.Context, review *Review) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetAverageRating(ctx context.Context, targetType string, targetID uuid.UUID) (float64, int, error)
}

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRepository creates a new reviews repository
func NewRepository(db *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

// Create creates a new review
func (r *RepositoryImpl) Create(ctx context.Context, review *Review) error {
	query := `
		INSERT INTO reviews (id, user_id, target_type, target_id, rating, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		review.ID,
		review.UserID,
		review.TargetType,
		review.TargetID,
		review.Rating,
		review.Comment,
		review.CreatedAt,
		review.UpdatedAt,
	)
	return err
}

// GetByID retrieves a review by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*Review, error) {
	query := `
		SELECT id, user_id, target_type, target_id, rating, comment, created_at, updated_at
		FROM reviews
		WHERE id = $1 AND deleted_at IS NULL
	`

	var review Review
	err := r.db.QueryRow(ctx, query, id).Scan(
		&review.ID,
		&review.UserID,
		&review.TargetType,
		&review.TargetID,
		&review.Rating,
		&review.Comment,
		&review.CreatedAt,
		&review.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &review, nil
}

// GetByUserAndTarget retrieves a review by user and target
func (r *RepositoryImpl) GetByUserAndTarget(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) (*Review, error) {
	query := `
		SELECT id, user_id, target_type, target_id, rating, comment, created_at, updated_at
		FROM reviews
		WHERE user_id = $1 AND target_type = $2 AND target_id = $3 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	var review Review
	err := r.db.QueryRow(ctx, query, userID, targetType, targetID).Scan(
		&review.ID,
		&review.UserID,
		&review.TargetType,
		&review.TargetID,
		&review.Rating,
		&review.Comment,
		&review.CreatedAt,
		&review.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &review, nil
}

// ListByTarget lists reviews for a target (restaurant, courier, or menu item)
func (r *RepositoryImpl) ListByTarget(ctx context.Context, targetType string, targetID uuid.UUID, limit int) ([]*Review, error) {
	query := `
		SELECT id, user_id, target_type, target_id, rating, comment, created_at, updated_at
		FROM reviews
		WHERE target_type = $1 AND target_id = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, query, targetType, targetID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*Review
	for rows.Next() {
		var review Review
		err := rows.Scan(
			&review.ID,
			&review.UserID,
			&review.TargetType,
			&review.TargetID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
			&review.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, &review)
	}

	return reviews, nil
}

// ListByUser lists reviews by a user
func (r *RepositoryImpl) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*Review, error) {
	query := `
		SELECT id, user_id, target_type, target_id, rating, comment, created_at, updated_at
		FROM reviews
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*Review
	for rows.Next() {
		var review Review
		err := rows.Scan(
			&review.ID,
			&review.UserID,
			&review.TargetType,
			&review.TargetID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
			&review.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, &review)
	}

	return reviews, nil
}

// Update updates a review
func (r *RepositoryImpl) Update(ctx context.Context, review *Review) error {
	query := `
		UPDATE reviews
		SET rating = $2, comment = $3, updated_at = $4
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, review.ID, review.Rating, review.Comment, review.UpdatedAt)
	return err
}

// Delete deletes a review (soft delete)
func (r *RepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE reviews
		SET deleted_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	return err
}

// GetAverageRating calculates the average rating for a target
func (r *RepositoryImpl) GetAverageRating(ctx context.Context, targetType string, targetID uuid.UUID) (float64, int, error) {
	query := `
		SELECT AVG(rating) as avg_rating, COUNT(*) as count
		FROM reviews
		WHERE target_type = $1 AND target_id = $2 AND deleted_at IS NULL
	`

	var avgRating float64
	var count int
	err := r.db.QueryRow(ctx, query, targetType, targetID).Scan(&avgRating, &count)
	if err != nil {
		return 0, 0, err
	}

	return avgRating, count, nil
}
