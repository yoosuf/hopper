package jobs

import (
	"context"
	"log"
)

// Service handles job scheduling and management
type Service struct {
	workerJobQueue chan<- interface{}
}

// New creates a new jobs service
func New(workerJobQueue chan<- interface{}) *Service {
	return &Service{
		workerJobQueue: workerJobQueue,
	}
}

// EnqueueOrderJob enqueues an order-related job
func (s *Service) EnqueueOrderJob(ctx context.Context, orderID string, jobType string) {
	log.Printf("Enqueuing order job: %s (type: %s)", orderID, jobType)
	job := map[string]interface{}{
		"id":   orderID,
		"type": jobType,
	}
	s.workerJobQueue <- job
}

// EnqueuePaymentJob enqueues a payment-related job
func (s *Service) EnqueuePaymentJob(ctx context.Context, paymentID string, jobType string) {
	log.Printf("Enqueuing payment job: %s (type: %s)", paymentID, jobType)
	job := map[string]interface{}{
		"id":   paymentID,
		"type": jobType,
	}
	s.workerJobQueue <- job
}

// EnqueueDeliveryJob enqueues a delivery-related job
func (s *Service) EnqueueDeliveryJob(ctx context.Context, deliveryID string, jobType string) {
	log.Printf("Enqueuing delivery job: %s (type: %s)", deliveryID, jobType)
	job := map[string]interface{}{
		"id":   deliveryID,
		"type": jobType,
	}
	s.workerJobQueue <- job
}
