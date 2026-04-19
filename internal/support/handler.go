package support

import (
	"encoding/json"
	"net/http"

	"github.com/yoosuf/hopper/internal/platform/httpx"
	"github.com/yoosuf/hopper/internal/platform/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles support HTTP requests
type Handler struct {
	service   *Service
	validator *validator.Validator
}

// NewHandler creates a new support handler
func NewHandler(service *Service, validator *validator.Validator) *Handler {
	return &Handler{
		service:   service,
		validator: validator,
	}
}

// CreateTicket handles creating a new support ticket
// @Summary Create a support ticket
// @Description Create a new customer support ticket
// @Tags support
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateTicketRequest true "Ticket details"
// @Success 201 {object} Ticket "Ticket created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /support/tickets [post]
func (h *Handler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	var req CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err)
		return
	}

	// Get user ID from context
	userIDStr := r.Context().Value("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID", nil)
		return
	}

	ticket, err := h.service.CreateTicket(r.Context(), userID, &req)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "CREATE_TICKET_ERROR", "Failed to create ticket", err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, ticket)
}

// GetTicket handles retrieving a ticket by ID
// @Summary Get support ticket by ID
// @Description Retrieve a specific support ticket by its ID
// @Tags support
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Ticket ID"
// @Success 200 {object} Ticket "Ticket retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid ticket ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Ticket not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /support/tickets/{id} [get]
func (h *Handler) GetTicket(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid ticket ID", nil)
		return
	}

	ticket, err := h.service.GetTicket(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Ticket not found", err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, ticket)
}

// ListUserTickets handles listing tickets for the current user
// @Summary List user support tickets
// @Description Retrieve all support tickets for the authenticated user
// @Tags support
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Ticket "Tickets retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /support/tickets [get]
func (h *Handler) ListUserTickets(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID", nil)
		return
	}

	tickets, err := h.service.ListUserTickets(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "LIST_TICKETS_ERROR", "Failed to list tickets", err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, tickets)
}

// UpdateTicket handles updating a ticket
// @Summary Update support ticket
// @Description Update a support ticket (status, priority, assignment)
// @Tags support
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Ticket ID"
// @Param request body UpdateTicketRequest true "Ticket update details"
// @Success 200 {object} Ticket "Ticket updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Ticket not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /support/tickets/{id} [put]
func (h *Handler) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid ticket ID", nil)
		return
	}

	var req UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err)
		return
	}

	err = h.service.UpdateTicket(r.Context(), id, &req)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "UPDATE_TICKET_ERROR", "Failed to update ticket", err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "Ticket updated successfully"})
}

// CreateMessage handles creating a new message
// @Summary Create message in support ticket
// @Description Add a message to a support ticket
// @Tags support
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Ticket ID"
// @Param request body CreateMessageRequest true "Message details"
// @Success 201 {object} Message "Message created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Ticket not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /support/tickets/{id}/messages [post]
func (h *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	ticketIDStr := chi.URLParam(r, "id")
	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid ticket ID", nil)
		return
	}

	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err)
		return
	}

	userIDStr := r.Context().Value("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID", nil)
		return
	}

	message, err := h.service.CreateMessage(r.Context(), ticketID, userID, false, &req)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "CREATE_MESSAGE_ERROR", "Failed to create message", err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, message)
}

// ListMessages handles listing messages for a ticket
// @Summary List messages in support ticket
// @Description Retrieve all messages for a specific support ticket
// @Tags support
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Ticket ID"
// @Success 200 {array} Message "Messages retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid ticket ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Ticket not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /support/tickets/{id}/messages [get]
func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	ticketIDStr := chi.URLParam(r, "id")
	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid ticket ID", nil)
		return
	}

	messages, err := h.service.ListMessages(r.Context(), ticketID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "LIST_MESSAGES_ERROR", "Failed to list messages", err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, messages)
}
