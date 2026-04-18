package support

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TicketStatus represents the status of a support ticket
type TicketStatus string

const (
	TicketStatusOpen      TicketStatus = "open"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusResolved  TicketStatus = "resolved"
	TicketStatusClosed    TicketStatus = "closed"
)

// TicketPriority represents the priority of a support ticket
type TicketPriority string

const (
	TicketPriorityLow    TicketPriority = "low"
	TicketPriorityMedium TicketPriority = "medium"
	TicketPriorityHigh   TicketPriority = "high"
	TicketPriorityUrgent TicketPriority = "urgent"
)

// Ticket represents a support ticket
type Ticket struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Subject     string
	Description string
	Category    string
	Status      TicketStatus
	Priority    TicketPriority
	AssignedTo  *uuid.UUID
	ResolvedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Message represents a support ticket message
type Message struct {
	ID        uuid.UUID
	TicketID  uuid.UUID
	UserID    uuid.UUID
	Content   string
	IsSupport bool
	CreatedAt time.Time
}

// Repository defines the interface for support data access
type Repository interface {
	// Ticket operations
	CreateTicket(ctx context.Context, ticket *Ticket) error
	GetTicketByID(ctx context.Context, id uuid.UUID) (*Ticket, error)
	ListTicketsByUser(ctx context.Context, userID uuid.UUID) ([]*Ticket, error)
	ListTicketsByStatus(ctx context.Context, status TicketStatus) ([]*Ticket, error)
	UpdateTicketStatus(ctx context.Context, id uuid.UUID, status TicketStatus) error
	AssignTicket(ctx context.Context, ticketID, assignedTo uuid.UUID) error
	
	// Message operations
	CreateMessage(ctx context.Context, message *Message) error
	ListMessagesByTicket(ctx context.Context, ticketID uuid.UUID) ([]*Message, error)
}

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRepository creates a new support repository
func NewRepository(db *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

// CreateTicket creates a new support ticket
func (r *RepositoryImpl) CreateTicket(ctx context.Context, ticket *Ticket) error {
	query := `
		INSERT INTO support_tickets (id, user_id, subject, description, category, status, priority, assigned_to, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(ctx, query,
		ticket.ID,
		ticket.UserID,
		ticket.Subject,
		ticket.Description,
		ticket.Category,
		ticket.Status,
		ticket.Priority,
		ticket.AssignedTo,
		ticket.CreatedAt,
		ticket.UpdatedAt,
	)
	return err
}

// GetTicketByID retrieves a ticket by ID
func (r *RepositoryImpl) GetTicketByID(ctx context.Context, id uuid.UUID) (*Ticket, error) {
	query := `
		SELECT id, user_id, subject, description, category, status, priority, assigned_to, resolved_at, created_at, updated_at
		FROM support_tickets
		WHERE id = $1 AND deleted_at IS NULL
	`

	var ticket Ticket
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ticket.ID,
		&ticket.UserID,
		&ticket.Subject,
		&ticket.Description,
		&ticket.Category,
		&ticket.Status,
		&ticket.Priority,
		&ticket.AssignedTo,
		&ticket.ResolvedAt,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &ticket, nil
}

// ListTicketsByUser lists tickets for a user
func (r *RepositoryImpl) ListTicketsByUser(ctx context.Context, userID uuid.UUID) ([]*Ticket, error) {
	query := `
		SELECT id, user_id, subject, description, category, status, priority, assigned_to, resolved_at, created_at, updated_at
		FROM support_tickets
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*Ticket
	for rows.Next() {
		var ticket Ticket
		err := rows.Scan(
			&ticket.ID,
			&ticket.UserID,
			&ticket.Subject,
			&ticket.Description,
			&ticket.Category,
			&ticket.Status,
			&ticket.Priority,
			&ticket.AssignedTo,
			&ticket.ResolvedAt,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, &ticket)
	}

	return tickets, nil
}

// ListTicketsByStatus lists tickets by status
func (r *RepositoryImpl) ListTicketsByStatus(ctx context.Context, status TicketStatus) ([]*Ticket, error) {
	query := `
		SELECT id, user_id, subject, description, category, status, priority, assigned_to, resolved_at, created_at, updated_at
		FROM support_tickets
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*Ticket
	for rows.Next() {
		var ticket Ticket
		err := rows.Scan(
			&ticket.ID,
			&ticket.UserID,
			&ticket.Subject,
			&ticket.Description,
			&ticket.Category,
			&ticket.Status,
			&ticket.Priority,
			&ticket.AssignedTo,
			&ticket.ResolvedAt,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, &ticket)
	}

	return tickets, nil
}

// UpdateTicketStatus updates a ticket's status
func (r *RepositoryImpl) UpdateTicketStatus(ctx context.Context, id uuid.UUID, status TicketStatus) error {
	query := `
		UPDATE support_tickets
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	now := time.Now().UTC()
	_, err := r.db.Exec(ctx, query, id, status, now)
	return err
}

// AssignTicket assigns a ticket to a support agent
func (r *RepositoryImpl) AssignTicket(ctx context.Context, ticketID, assignedTo uuid.UUID) error {
	query := `
		UPDATE support_tickets
		SET assigned_to = $2, updated_at = $3
		WHERE id = $1
	`

	now := time.Now().UTC()
	_, err := r.db.Exec(ctx, query, ticketID, assignedTo, now)
	return err
}

// CreateMessage creates a new message
func (r *RepositoryImpl) CreateMessage(ctx context.Context, message *Message) error {
	query := `
		INSERT INTO support_messages (id, ticket_id, user_id, content, is_support, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, query,
		message.ID,
		message.TicketID,
		message.UserID,
		message.Content,
		message.IsSupport,
		message.CreatedAt,
	)
	return err
}

// ListMessagesByTicket lists messages for a ticket
func (r *RepositoryImpl) ListMessagesByTicket(ctx context.Context, ticketID uuid.UUID) ([]*Message, error) {
	query := `
		SELECT id, ticket_id, user_id, content, is_support, created_at
		FROM support_messages
		WHERE ticket_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var message Message
		err := rows.Scan(
			&message.ID,
			&message.TicketID,
			&message.UserID,
			&message.Content,
			&message.IsSupport,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}

	return messages, nil
}
