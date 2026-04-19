package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yoosuf/hopper/internal/platform/logger"
)

// Service handles worker operations
type Service struct {
	jobQueue   chan Job
	logger     logger.Logger
	workerPool chan struct{}
	workers    int
}

// New creates a new worker service
func New(jobQueue chan Job, log logger.Logger, workers int) *Service {
	if workers <= 0 {
		workers = 5 // Default worker count
	}

	return &Service{
		jobQueue:   jobQueue,
		logger:     log,
		workerPool: make(chan struct{}, workers),
		workers:    workers,
	}
}

// Job represents a background job
type Job struct {
	ID         string
	Type       string
	Data       interface{}
	RetryCount int
	MaxRetries int
}

// JobType represents the type of job
type JobType string

const (
	JobTypeOrderCreated      JobType = "order_created"
	JobTypeOrderConfirmed    JobType = "order_confirmed"
	JobTypeOrderPreparing    JobType = "order_preparing"
	JobTypeOrderReady        JobType = "order_ready"
	JobTypeOrderPickedUp     JobType = "order_picked_up"
	JobTypeOrderDelivered    JobType = "order_delivered"
	JobTypeOrderCancelled    JobType = "order_cancelled"
	JobTypePaymentReceived   JobType = "payment_received"
	JobTypePaymentFailed     JobType = "payment_failed"
	JobTypeDeliveryAssigned  JobType = "delivery_assigned"
	JobTypeDeliveryCompleted JobType = "delivery_completed"
	JobTypeNotificationSend  JobType = "notification_send"
)

// JobData represents job-specific data
type JobData struct {
	OrderID          *uuid.UUID
	PaymentID        *uuid.UUID
	DeliveryID       *uuid.UUID
	UserID           *uuid.UUID
	NotificationData map[string]interface{}
}

// Start starts the worker pool
func (s *Service) Start(ctx context.Context) {
	s.logger.Info("Worker pool started", logger.F("workers", s.workers))

	// Start worker goroutines
	for i := 0; i < s.workers; i++ {
		go s.worker(ctx)
	}

	// Wait for context cancellation
	<-ctx.Done()
	s.logger.Info("Worker pool stopped")
}

// worker is a single worker goroutine that processes jobs
func (s *Service) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-s.jobQueue:
			s.processJob(ctx, job)
		}
	}
}

// processJob processes a single job
func (s *Service) processJob(ctx context.Context, job Job) {
	s.logger.Info("Processing job", logger.F("job_id", job.ID), logger.F("job_type", job.Type))

	var err error
	switch JobType(job.Type) {
	case JobTypeOrderCreated:
		err = s.handleOrderCreated(ctx, job)
	case JobTypeOrderConfirmed:
		err = s.handleOrderConfirmed(ctx, job)
	case JobTypeOrderPreparing:
		err = s.handleOrderPreparing(ctx, job)
	case JobTypeOrderReady:
		err = s.handleOrderReady(ctx, job)
	case JobTypeOrderPickedUp:
		err = s.handleOrderPickedUp(ctx, job)
	case JobTypeOrderDelivered:
		err = s.handleOrderDelivered(ctx, job)
	case JobTypeOrderCancelled:
		err = s.handleOrderCancelled(ctx, job)
	case JobTypePaymentReceived:
		err = s.handlePaymentReceived(ctx, job)
	case JobTypePaymentFailed:
		err = s.handlePaymentFailed(ctx, job)
	case JobTypeDeliveryAssigned:
		err = s.handleDeliveryAssigned(ctx, job)
	case JobTypeDeliveryCompleted:
		err = s.handleDeliveryCompleted(ctx, job)
	case JobTypeNotificationSend:
		err = s.handleNotificationSend(ctx, job)
	default:
		s.logger.Warn("Unknown job type", logger.F("job_type", job.Type))
	}

	if err != nil {
		s.logger.Error("Error processing job", logger.F("job_id", job.ID), logger.F("error", err))
		s.handleJobError(ctx, job, err)
	} else {
		s.logger.Info("Successfully processed job", logger.F("job_id", job.ID))
	}
}

// handleJobError handles job processing errors with retry logic
func (s *Service) handleJobError(ctx context.Context, job Job, err error) {
	if job.MaxRetries == 0 {
		job.MaxRetries = 3 // Default max retries
	}

	if job.RetryCount < job.MaxRetries {
		job.RetryCount++
		retryDelay := time.Duration(job.RetryCount) * time.Second
		s.logger.Info("Retrying job", logger.F("job_id", job.ID), logger.F("retry_count", job.RetryCount), logger.F("max_retries", job.MaxRetries), logger.F("retry_delay", retryDelay))

		go func() {
			time.Sleep(retryDelay)
			select {
			case s.jobQueue <- job:
			case <-ctx.Done():
				s.logger.Info("Retry cancelled for job", logger.F("job_id", job.ID))
			}
		}()
	} else {
		s.logger.Error("Job failed after max retries", logger.F("job_id", job.ID), logger.F("max_retries", job.MaxRetries))
		// TODO: Implement dead letter queue for failed jobs
	}
}

// handleOrderCreated handles order created job
func (s *Service) handleOrderCreated(ctx context.Context, job Job) error {
	data, err := s.parseJobData(job.Data)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	s.logger.Info("Order created job", logger.F("order_id", data.OrderID))

	// TODO:
	// - Send confirmation notification to customer
	// - Send notification to restaurant
	// - Trigger payment processing
	// - Assign delivery courier

	return nil
}

// handleOrderConfirmed handles order confirmed job
func (s *Service) handleOrderConfirmed(ctx context.Context, job Job) error {
	data, err := s.parseJobData(job.Data)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	s.logger.Info("Order confirmed job", logger.F("order_id", data.OrderID))

	// TODO:
	// - Send notification to customer
	// - Update order status

	return nil
}

// handleOrderPreparing handles order preparing job
func (s *Service) handleOrderPreparing(ctx context.Context, job Job) error {
	data, err := s.parseJobData(job.Data)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	s.logger.Info("Order preparing job", logger.F("order_id", data.OrderID))

	// TODO:
	// - Send notification to customer
	// - Update order status

	return nil
}

// handleOrderReady handles order ready job
func (s *Service) handleOrderReady(ctx context.Context, job Job) error {
	data, err := s.parseJobData(job.Data)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	s.logger.Info("Order ready job", logger.F("order_id", data.OrderID))

	// TODO:
	// - Send notification to customer
	// - Notify courier
	// - Update order status

	return nil
}

// handleOrderPickedUp handles order picked up job
func (s *Service) handleOrderPickedUp(ctx context.Context, job Job) error {
	data, err := s.parseJobData(job.Data)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	s.logger.Info("Order picked up job", logger.F("order_id", data.OrderID))

	// TODO:
	// - Send notification to customer
	// - Update order status
	// - Track delivery progress

	return nil
}

// handleOrderDelivered handles order delivered job
func (s *Service) handleOrderDelivered(ctx context.Context, job Job) error {
	data, err := s.parseJobData(job.Data)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	s.logger.Info("Order delivered job", logger.F("order_id", data.OrderID))

	// TODO:
	// - Send notification to customer
	// - Update order status
	// - Process delivery completion
	// - Trigger review request

	return nil
}

// handleOrderCancelled handles order cancelled job
func (s *Service) handleOrderCancelled(ctx context.Context, job Job) error {
	data, err := s.parseJobData(job.Data)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	s.logger.Info("Order cancelled job", logger.F("order_id", data.OrderID))

	// TODO:
	// - Send notification to customer
	// - Send notification to restaurant
	// - Process refund if applicable
	// - Cancel delivery assignment

	return nil
}

// handlePaymentReceived handles payment received job
func (s *Service) handlePaymentReceived(ctx context.Context, job Job) error {
	data, err := s.parseJobData(job.Data)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	s.logger.Info("Payment received job", logger.F("payment_id", data.PaymentID))

	// TODO:
	// - Update payment status
	// - Confirm order
	// - Send notification

	return nil
}

// handlePaymentFailed handles payment failed job
func (s *Service) handlePaymentFailed(ctx context.Context, job Job) error {
	data, err := s.parseJobData(job.Data)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	s.logger.Warn("Payment failed job", logger.F("payment_id", data.PaymentID))

	// TODO:
	// - Update payment status
	// - Cancel order
	// - Send notification to customer

	return nil
}

// handleDeliveryAssigned handles delivery assigned job
func (s *Service) handleDeliveryAssigned(ctx context.Context, job Job) error {
	data, err := s.parseJobData(job.Data)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	s.logger.Info("Delivery assigned job", logger.F("delivery_id", data.DeliveryID))

	// TODO:
	// - Notify courier
	// - Update delivery status
	// - Send notification to customer

	return nil
}

// handleDeliveryCompleted handles delivery completed job
func (s *Service) handleDeliveryCompleted(ctx context.Context, job Job) error {
	data, err := s.parseJobData(job.Data)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	s.logger.Info("Delivery completed job", logger.F("delivery_id", data.DeliveryID))

	// TODO:
	// - Update delivery status
	// - Process courier payment
	// - Send notification

	return nil
}

// handleNotificationSend handles notification send job
func (s *Service) handleNotificationSend(ctx context.Context, job Job) error {
	data, err := s.parseJobData(job.Data)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	s.logger.Info("Notification send job", logger.F("user_id", data.UserID))

	// TODO:
	// - Send push notification
	// - Send email notification
	// - Send SMS notification
	// - Mark notification as sent

	return nil
}

// parseJobData parses job data into JobData struct
func (s *Service) parseJobData(data interface{}) (*JobData, error) {
	var jobData JobData

	// Try to unmarshal if data is a string/bytes
	switch v := data.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &jobData); err != nil {
			return nil, err
		}
	case []byte:
		if err := json.Unmarshal(v, &jobData); err != nil {
			return nil, err
		}
	case map[string]interface{}:
		// Handle map data
		if orderID, ok := v["order_id"].(string); ok {
			id, err := uuid.Parse(orderID)
			if err != nil {
				return nil, err
			}
			jobData.OrderID = &id
		}
		if paymentID, ok := v["payment_id"].(string); ok {
			id, err := uuid.Parse(paymentID)
			if err != nil {
				return nil, err
			}
			jobData.PaymentID = &id
		}
		if deliveryID, ok := v["delivery_id"].(string); ok {
			id, err := uuid.Parse(deliveryID)
			if err != nil {
				return nil, err
			}
			jobData.DeliveryID = &id
		}
		if userID, ok := v["user_id"].(string); ok {
			id, err := uuid.Parse(userID)
			if err != nil {
				return nil, err
			}
			jobData.UserID = &id
		}
		if notificationData, ok := v["notification_data"].(map[string]interface{}); ok {
			jobData.NotificationData = notificationData
		}
	default:
		return nil, fmt.Errorf("unsupported job data type: %T", data)
	}

	return &jobData, nil
}
