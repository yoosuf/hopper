package support

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles support operations
type Service struct {
	repo Repository
}

// New creates a new support service
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateTicketRequest represents a create ticket request
type CreateTicketRequest struct {
	Subject     string
	Description string
	Category    string
	Priority    TicketPriority
}

// UpdateTicketRequest represents an update ticket request
type UpdateTicketRequest struct {
	Status   *TicketStatus
	Priority *TicketPriority
}

// CreateTicket creates a new support ticket
func (s *Service) CreateTicket(ctx context.Context, userID uuid.UUID, req *CreateTicketRequest) (*Ticket, error) {
	ticket := &Ticket{
		ID:          uuid.New(),
		UserID:      userID,
		Subject:     req.Subject,
		Description: req.Description,
		Category:    req.Category,
		Status:      TicketStatusOpen,
		Priority:    req.Priority,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := s.repo.CreateTicket(ctx, ticket); err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	return ticket, nil
}

// GetTicket retrieves a ticket by ID
func (s *Service) GetTicket(ctx context.Context, id uuid.UUID) (*Ticket, error) {
	return s.repo.GetTicketByID(ctx, id)
}

// ListUserTickets lists tickets for a user
func (s *Service) ListUserTickets(ctx context.Context, userID uuid.UUID) ([]*Ticket, error) {
	return s.repo.ListTicketsByUser(ctx, userID)
}

// ListTicketsByStatus lists tickets by status
func (s *Service) ListTicketsByStatus(ctx context.Context, status TicketStatus) ([]*Ticket, error) {
	return s.repo.ListTicketsByStatus(ctx, status)
}

// UpdateTicket updates a ticket
func (s *Service) UpdateTicket(ctx context.Context, id uuid.UUID, req *UpdateTicketRequest) error {
	if req.Status != nil {
		if err := s.repo.UpdateTicketStatus(ctx, id, *req.Status); err != nil {
			return fmt.Errorf("failed to update ticket status: %w", err)
		}
	}
	return nil
}

// AssignTicket assigns a ticket to a support agent
func (s *Service) AssignTicket(ctx context.Context, ticketID, assignedTo uuid.UUID) error {
	return s.repo.AssignTicket(ctx, ticketID, assignedTo)
}

// CreateMessageRequest represents a create message request
type CreateMessageRequest struct {
	Content string
}

// CreateMessage creates a new message
func (s *Service) CreateMessage(ctx context.Context, ticketID, userID uuid.UUID, isSupport bool, req *CreateMessageRequest) (*Message, error) {
	message := &Message{
		ID:        uuid.New(),
		TicketID:  ticketID,
		UserID:    userID,
		Content:   req.Content,
		IsSupport: isSupport,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.CreateMessage(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return message, nil
}

// ListMessages lists messages for a ticket
func (s *Service) ListMessages(ctx context.Context, ticketID uuid.UUID) ([]*Message, error) {
	return s.repo.ListMessagesByTicket(ctx, ticketID)
}
