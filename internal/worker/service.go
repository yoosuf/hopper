package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// Service handles worker operations
type Service struct {
	jobQueue chan Job
}

// New creates a new worker service
func New(jobQueue chan Job) *Service {
	return &Service{
		jobQueue: jobQueue,
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

// Start starts the worker
func (s *Service) Start(ctx context.Context) {
	log.Println("Worker started")
	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopped")
			return
		case job := <-s.jobQueue:
			s.processJob(ctx, job)
		}
	}
}

// processJob processes a single job
func (s *Service) processJob(ctx context.Context, job Job) {
	log.Printf("Processing job: %s (type: %s)", job.ID, job.Type)

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
		log.Printf("Unknown job type: %s", job.Type)
		return
	}

	if err != nil {
		log.Printf("Error processing job %s: %v", job.ID, err)
		s.handleJobError(ctx, job, err)
	} else {
		log.Printf("Successfully processed job: %s", job.ID)
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
		log.Printf("Retrying job %s (attempt %d/%d) in %v", job.ID, job.RetryCount, job.MaxRetries, retryDelay)

		go func() {
			time.Sleep(retryDelay)
			select {
			case s.jobQueue <- job:
			case <-ctx.Done():
				log.Printf("Retry cancelled for job %s", job.ID)
			}
		}()
	} else {
		log.Printf("Job %s failed after %d retries", job.ID, job.MaxRetries)
		// TODO: Implement dead letter queue for failed jobs
	}
}

// handleOrderCreated handles order created job
func (s *Service) handleOrderCreated(ctx context.Context, job Job) error {
	data, err := s.parseJobData(job.Data)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	log.Printf("Order created job: OrderID=%v", data.OrderID)

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

	log.Printf("Order confirmed job: OrderID=%v", data.OrderID)

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

	log.Printf("Order preparing job: OrderID=%v", data.OrderID)

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

	log.Printf("Order ready job: OrderID=%v", data.OrderID)

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

	log.Printf("Order picked up job: OrderID=%v", data.OrderID)

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

	log.Printf("Order delivered job: OrderID=%v", data.OrderID)

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

	log.Printf("Order cancelled job: OrderID=%v", data.OrderID)

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

	log.Printf("Payment received job: PaymentID=%v", data.PaymentID)

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

	log.Printf("Payment failed job: PaymentID=%v", data.PaymentID)

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

	log.Printf("Delivery assigned job: DeliveryID=%v", data.DeliveryID)

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

	log.Printf("Delivery completed job: DeliveryID=%v", data.DeliveryID)

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

	log.Printf("Notification send job: UserID=%v", data.UserID)

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
