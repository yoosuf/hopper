package jobs

import (
	"context"

	"github.com/yoosuf/hopper/internal/platform/logger"
)

// Service handles job scheduling and management
type Service struct {
	workerJobQueue chan<- interface{}
	logger         logger.Logger
}

// New creates a new jobs service
func New(workerJobQueue chan<- interface{}, log logger.Logger) *Service {
	return &Service{
		workerJobQueue: workerJobQueue,
		logger:         log,
	}
}

// EnqueueOrderJob enqueues an order-related job
func (s *Service) EnqueueOrderJob(ctx context.Context, orderID string, jobType string) {
	s.logger.Info("Enqueuing order job", logger.F("order_id", orderID), logger.F("job_type", jobType))
	job := map[string]interface{}{
		"id":   orderID,
		"type": jobType,
	}
	s.workerJobQueue <- job
}

// EnqueuePaymentJob enqueues a payment-related job
func (s *Service) EnqueuePaymentJob(ctx context.Context, paymentID string, jobType string) {
	s.logger.Info("Enqueuing payment job", logger.F("payment_id", paymentID), logger.F("job_type", jobType))
	job := map[string]interface{}{
		"id":   paymentID,
		"type": jobType,
	}
	s.workerJobQueue <- job
}

// EnqueueDeliveryJob enqueues a delivery-related job
func (s *Service) EnqueueDeliveryJob(ctx context.Context, deliveryID string, jobType string) {
	s.logger.Info("Enqueuing delivery job", logger.F("delivery_id", deliveryID), logger.F("job_type", jobType))
	job := map[string]interface{}{
		"id":   deliveryID,
		"type": jobType,
	}
	s.workerJobQueue <- job
}
