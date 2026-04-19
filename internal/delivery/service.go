package delivery

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/yoosuf/hopper/internal/platform/logger"
)

// Service handles delivery operations
type Service struct {
	repo     Repository
	flags    FeatureFlags
	log      logger.Logger
	maps     MapsProvider
	alerting AlertingProvider
	notifier CourierNotifier
}

// FeatureFlags controls courier automation capabilities.
type FeatureFlags struct {
	AutoDispatchEnabled         bool
	RouteOptimizationEnabled    bool
	LiveTrackingEnabled         bool
	AutoReassignEnabled         bool
	SLAMonitoringEnabled        bool
	ProviderIntegrationsEnabled bool
	DispatchRadiusKM            float64
	ReassignTimeout             time.Duration
	SLAThreshold                time.Duration
	AverageSpeedKPH             float64
}

// MapsProvider abstracts route and ETA integrations.
type MapsProvider interface {
	EstimateRoute(ctx context.Context, from, to Location) (*RouteEstimate, error)
}

// AlertingProvider abstracts SLA alerting integration.
type AlertingProvider interface {
	SendSLAAlert(ctx context.Context, deliveryID uuid.UUID, message string) error
}

// CourierNotifier abstracts courier/customer delivery notifications.
type CourierNotifier interface {
	NotifyCourierAssigned(ctx context.Context, courierID, deliveryID uuid.UUID, etaMinutes int) error
}

// RouteEstimate represents route optimization output.
type RouteEstimate struct {
	DistanceKM float64
	ETAMinutes int
}

// New creates a new delivery service
func New(repo Repository, flags FeatureFlags, log logger.Logger, maps MapsProvider, alerting AlertingProvider, notifier CourierNotifier) *Service {
	if flags.DispatchRadiusKM <= 0 {
		flags.DispatchRadiusKM = 10
	}
	if flags.ReassignTimeout <= 0 {
		flags.ReassignTimeout = 3 * time.Minute
	}
	if flags.SLAThreshold <= 0 {
		flags.SLAThreshold = 15 * time.Minute
	}
	if flags.AverageSpeedKPH <= 0 {
		flags.AverageSpeedKPH = 25
	}

	return &Service{repo: repo, flags: flags, log: log, maps: maps, alerting: alerting, notifier: notifier}
}

// Delivery represents a delivery assignment
type Delivery struct {
	ID                  uuid.UUID
	OrderID             uuid.UUID
	CourierID           *uuid.UUID
	Status              string
	PickupLocation      Location
	DeliveryLocation    Location
	EstimatedDeliveryAt *time.Time
	AssignedAt          *time.Time
	PickedUpAt          *time.Time
	DeliveredAt         *time.Time
	CreatedAt           time.Time
}

// Location represents a geographic location
type Location struct {
	Latitude  float64
	Longitude float64
}

// DeliveryFeeRequest represents a delivery fee calculation request
type DeliveryFeeRequest struct {
	PickupLocation   Location
	DeliveryLocation Location
	BaseFee          int
	DistanceFeePerKm int
	SurgeMultiplier  float64
	IsSurgeActive    bool
}

// DeliveryFeeResponse represents a delivery fee calculation response
type DeliveryFeeRequestResponse struct {
	BaseFee         int
	DistanceKm      float64
	DistanceFee     int
	Surcharge       int
	TotalFee        int
	SurgeMultiplier float64
	IsSurgeActive   bool
}

// CourierCandidate represents an available courier for assignment.
type CourierCandidate struct {
	CourierID uuid.UUID
	Location  Location
}

// CreateDelivery creates a new delivery assignment
func (s *Service) CreateDelivery(ctx context.Context, orderID uuid.UUID) (*Delivery, error) {
	delivery := &Delivery{
		ID:      uuid.New(),
		OrderID: orderID,
		Status:  "unassigned",
	}

	if err := s.repo.Create(ctx, delivery); err != nil {
		return nil, fmt.Errorf("failed to create delivery: %w", err)
	}

	return delivery, nil
}

// GetDelivery retrieves a delivery by ID
func (s *Service) GetDelivery(ctx context.Context, id uuid.UUID) (*Delivery, error) {
	return s.repo.GetByID(ctx, id)
}

// ListCourierDeliveries lists deliveries for a courier
func (s *Service) ListCourierDeliveries(ctx context.Context, courierID uuid.UUID) ([]*Delivery, error) {
	return s.repo.ListByCourier(ctx, courierID)
}

// AssignCourier assigns a courier to a delivery
func (s *Service) AssignCourier(ctx context.Context, deliveryID uuid.UUID, courierID uuid.UUID) error {
	return s.repo.AssignCourier(ctx, deliveryID, courierID)
}

// UpdateStatus updates a delivery's status
func (s *Service) UpdateStatus(ctx context.Context, deliveryID uuid.UUID, status string) error {
	return s.repo.UpdateStatus(ctx, deliveryID, status)
}

// MarkPickedUp marks a delivery as picked up
func (s *Service) MarkPickedUp(ctx context.Context, deliveryID uuid.UUID) error {
	return s.repo.MarkPickedUp(ctx, deliveryID)
}

// MarkDelivered marks a delivery as delivered
func (s *Service) MarkDelivered(ctx context.Context, deliveryID uuid.UUID) error {
	return s.repo.MarkDelivered(ctx, deliveryID)
}

// AutoDispatch finds and assigns the best courier for a delivery when enabled.
func (s *Service) AutoDispatch(ctx context.Context, deliveryID uuid.UUID) (*uuid.UUID, error) {
	if !s.flags.AutoDispatchEnabled {
		return nil, nil
	}

	delivery, err := s.repo.GetByID(ctx, deliveryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery for dispatch: %w", err)
	}

	couriers, err := s.repo.ListAvailableCouriers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list available couriers: %w", err)
	}

	bestCourierID, etaMinutes, found := s.selectBestCourier(ctx, delivery, couriers)
	if !found {
		s.log.Warn("No courier candidate found for delivery", logger.F("delivery_id", deliveryID))
		return nil, nil
	}

	if err := s.repo.AssignCourier(ctx, deliveryID, bestCourierID); err != nil {
		return nil, fmt.Errorf("failed to assign courier: %w", err)
	}

	if etaErr := s.repo.SetEstimatedDeliveryETA(ctx, deliveryID, etaMinutes); etaErr != nil {
		s.log.Warn("Failed to persist delivery ETA", logger.F("delivery_id", deliveryID), logger.F("error", etaErr))
	}

	if s.notifier != nil && s.flags.ProviderIntegrationsEnabled {
		if notifyErr := s.notifier.NotifyCourierAssigned(ctx, bestCourierID, deliveryID, etaMinutes); notifyErr != nil {
			s.log.Warn("Failed to send courier assignment notification", logger.F("delivery_id", deliveryID), logger.F("courier_id", bestCourierID), logger.F("error", notifyErr))
		}
	}

	return &bestCourierID, nil
}

// UpdateCourierLocation updates a courier's live location when enabled.
func (s *Service) UpdateCourierLocation(ctx context.Context, courierID uuid.UUID, latitude, longitude float64) error {
	if !s.flags.LiveTrackingEnabled {
		return nil
	}
	return s.repo.UpsertCourierLocation(ctx, courierID, latitude, longitude)
}

// AutoReassignTimedOut reassigns deliveries that are assigned for too long without progress.
func (s *Service) AutoReassignTimedOut(ctx context.Context) (int, error) {
	if !s.flags.AutoReassignEnabled {
		return 0, nil
	}

	candidates, err := s.repo.ListReassignmentCandidates(ctx, s.flags.ReassignTimeout)
	if err != nil {
		return 0, fmt.Errorf("failed to list reassignment candidates: %w", err)
	}

	reassigned := 0
	for _, delivery := range candidates {
		if delivery.CourierID != nil {
			_ = s.repo.UnassignCourier(ctx, delivery.ID)
		}
		assignedCourier, dispatchErr := s.AutoDispatch(ctx, delivery.ID)
		if dispatchErr != nil {
			s.log.Warn("Auto reassignment failed", logger.F("delivery_id", delivery.ID), logger.F("error", dispatchErr))
			continue
		}
		if assignedCourier != nil {
			reassigned++
		}
	}

	return reassigned, nil
}

// MonitorSLA checks delayed deliveries and sends alerts when enabled.
func (s *Service) MonitorSLA(ctx context.Context) (int, error) {
	if !s.flags.SLAMonitoringEnabled {
		return 0, nil
	}

	overdue, err := s.repo.ListDeliveriesPastSLA(ctx, s.flags.SLAThreshold)
	if err != nil {
		return 0, fmt.Errorf("failed to query SLA overdue deliveries: %w", err)
	}

	alerts := 0
	for _, delivery := range overdue {
		if s.alerting != nil && s.flags.ProviderIntegrationsEnabled {
			msg := fmt.Sprintf("Delivery %s breached SLA threshold", delivery.ID)
			if alertErr := s.alerting.SendSLAAlert(ctx, delivery.ID, msg); alertErr != nil {
				s.log.Warn("Failed to send SLA alert", logger.F("delivery_id", delivery.ID), logger.F("error", alertErr))
				continue
			}
		}
		alerts++
	}

	return alerts, nil
}

func (s *Service) selectBestCourier(ctx context.Context, delivery *Delivery, candidates []*CourierCandidate) (uuid.UUID, int, bool) {
	bestDistance := math.MaxFloat64
	var bestCourier uuid.UUID
	bestETA := 0

	for _, candidate := range candidates {
		distance := haversineKM(candidate.Location, delivery.PickupLocation)
		if distance > s.flags.DispatchRadiusKM {
			continue
		}

		pickupEstimate, err := s.estimateRoute(ctx, candidate.Location, delivery.PickupLocation)
		if err != nil {
			s.log.Warn("Failed route estimate for courier", logger.F("courier_id", candidate.CourierID), logger.F("error", err))
			continue
		}

		deliveryEstimate, err := s.estimateRoute(ctx, delivery.PickupLocation, delivery.DeliveryLocation)
		if err != nil {
			s.log.Warn("Failed route estimate for courier", logger.F("courier_id", candidate.CourierID), logger.F("error", err))
			continue
		}

		totalDistance := pickupEstimate.DistanceKM + deliveryEstimate.DistanceKM
		totalETA := pickupEstimate.ETAMinutes + deliveryEstimate.ETAMinutes

		if totalDistance < bestDistance {
			bestDistance = totalDistance
			bestCourier = candidate.CourierID
			bestETA = totalETA
		}
	}

	if bestDistance == math.MaxFloat64 {
		return uuid.Nil, 0, false
	}

	return bestCourier, bestETA, true
}

func (s *Service) estimateRoute(ctx context.Context, from, to Location) (*RouteEstimate, error) {
	if s.maps != nil && s.flags.ProviderIntegrationsEnabled && s.flags.RouteOptimizationEnabled {
		estimate, err := s.maps.EstimateRoute(ctx, from, to)
		if err == nil {
			return estimate, nil
		}
	}

	distanceKM := haversineKM(from, to)
	hours := distanceKM / s.flags.AverageSpeedKPH
	etaMinutes := int(math.Ceil(hours * 60))
	if etaMinutes < 1 {
		etaMinutes = 1
	}

	return &RouteEstimate{DistanceKM: distanceKM, ETAMinutes: etaMinutes}, nil
}

func haversineKM(a, b Location) float64 {
	const earthRadiusKM = 6371.0

	lat1 := a.Latitude * math.Pi / 180
	lat2 := b.Latitude * math.Pi / 180
	dLat := (b.Latitude - a.Latitude) * math.Pi / 180
	dLon := (b.Longitude - a.Longitude) * math.Pi / 180

	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))

	return earthRadiusKM * c
}
