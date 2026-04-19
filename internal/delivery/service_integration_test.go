package delivery

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yoosuf/hopper/internal/platform/logger"
)

type fakeRepository struct {
	deliveries         map[uuid.UUID]*Delivery
	availableCouriers  []*CourierCandidate
	reassignCandidates []*Delivery
	overdueDeliveries  []*Delivery
	etaByDeliveryID    map[uuid.UUID]int
	unassignCount      int
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		deliveries:      map[uuid.UUID]*Delivery{},
		etaByDeliveryID: map[uuid.UUID]int{},
	}
}

func (f *fakeRepository) Create(_ context.Context, delivery *Delivery) error {
	f.deliveries[delivery.ID] = delivery
	return nil
}

func (f *fakeRepository) GetByID(_ context.Context, id uuid.UUID) (*Delivery, error) {
	return f.deliveries[id], nil
}

func (f *fakeRepository) ListByCourier(_ context.Context, courierID uuid.UUID) ([]*Delivery, error) {
	result := make([]*Delivery, 0)
	for _, d := range f.deliveries {
		if d.CourierID != nil && *d.CourierID == courierID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (f *fakeRepository) ListAvailableCouriers(_ context.Context) ([]*CourierCandidate, error) {
	return f.availableCouriers, nil
}

func (f *fakeRepository) AssignCourier(_ context.Context, deliveryID uuid.UUID, courierID uuid.UUID) error {
	d := f.deliveries[deliveryID]
	d.CourierID = &courierID
	now := time.Now().UTC()
	d.AssignedAt = &now
	return nil
}

func (f *fakeRepository) SetEstimatedDeliveryETA(_ context.Context, deliveryID uuid.UUID, etaMinutes int) error {
	f.etaByDeliveryID[deliveryID] = etaMinutes
	return nil
}

func (f *fakeRepository) UnassignCourier(_ context.Context, deliveryID uuid.UUID) error {
	f.unassignCount++
	d := f.deliveries[deliveryID]
	d.CourierID = nil
	d.AssignedAt = nil
	return nil
}

func (f *fakeRepository) UpdateStatus(_ context.Context, deliveryID uuid.UUID, status string) error {
	f.deliveries[deliveryID].Status = status
	return nil
}

func (f *fakeRepository) UpsertCourierLocation(_ context.Context, _ uuid.UUID, _, _ float64) error {
	return nil
}

func (f *fakeRepository) ListReassignmentCandidates(_ context.Context, _ time.Duration) ([]*Delivery, error) {
	return f.reassignCandidates, nil
}

func (f *fakeRepository) ListDeliveriesPastSLA(_ context.Context, _ time.Duration) ([]*Delivery, error) {
	return f.overdueDeliveries, nil
}

func (f *fakeRepository) MarkPickedUp(_ context.Context, deliveryID uuid.UUID) error {
	f.deliveries[deliveryID].Status = "picked_up"
	return nil
}

func (f *fakeRepository) MarkDelivered(_ context.Context, deliveryID uuid.UUID) error {
	f.deliveries[deliveryID].Status = "delivered"
	now := time.Now().UTC()
	f.deliveries[deliveryID].DeliveredAt = &now
	return nil
}

type noopLogger struct{}

func (noopLogger) Debug(string, ...logger.Field)             {}
func (noopLogger) Info(string, ...logger.Field)              {}
func (noopLogger) Warn(string, ...logger.Field)              {}
func (noopLogger) Error(string, ...logger.Field)             {}
func (noopLogger) With(...logger.Field) logger.Logger        { return noopLogger{} }
func (noopLogger) WithContext(context.Context) logger.Logger { return noopLogger{} }

type alertingSpy struct {
	count int
}

func (a *alertingSpy) SendSLAAlert(_ context.Context, _ uuid.UUID, _ string) error {
	a.count++
	return nil
}

func TestAutoDispatchSelectsBestCourier(t *testing.T) {
	repo := newFakeRepository()
	deliveryID := uuid.New()
	repo.deliveries[deliveryID] = &Delivery{
		ID:               deliveryID,
		Status:           "unassigned",
		PickupLocation:   Location{Latitude: 25.2048, Longitude: 55.2708},
		DeliveryLocation: Location{Latitude: 25.1972, Longitude: 55.2744},
	}

	nearCourier := uuid.New()
	farCourier := uuid.New()
	repo.availableCouriers = []*CourierCandidate{
		{CourierID: farCourier, Location: Location{Latitude: 25.5000, Longitude: 55.6000}},
		{CourierID: nearCourier, Location: Location{Latitude: 25.2050, Longitude: 55.2710}},
	}

	svc := New(repo, FeatureFlags{
		AutoDispatchEnabled:      true,
		RouteOptimizationEnabled: false,
		DispatchRadiusKM:         50,
		AverageSpeedKPH:          25,
	}, noopLogger{}, NewMockMapsProvider(25), nil, nil)

	assigned, err := svc.AutoDispatch(context.Background(), deliveryID)
	if err != nil {
		t.Fatalf("AutoDispatch returned error: %v", err)
	}
	if assigned == nil {
		t.Fatalf("expected a courier assignment")
	}
	if *assigned != nearCourier {
		t.Fatalf("expected nearest courier %s, got %s", nearCourier, *assigned)
	}
	if repo.deliveries[deliveryID].CourierID == nil || *repo.deliveries[deliveryID].CourierID != nearCourier {
		t.Fatalf("expected repository courier assignment to be persisted")
	}
	if repo.etaByDeliveryID[deliveryID] <= 0 {
		t.Fatalf("expected ETA to be stored")
	}
}

func TestAutoReassignTimedOutReassignsDelivery(t *testing.T) {
	repo := newFakeRepository()
	deliveryID := uuid.New()
	oldCourier := uuid.New()
	newCourier := uuid.New()
	now := time.Now().UTC().Add(-10 * time.Minute)

	repo.deliveries[deliveryID] = &Delivery{
		ID:               deliveryID,
		Status:           "assigned",
		CourierID:        &oldCourier,
		AssignedAt:       &now,
		PickupLocation:   Location{Latitude: 25.2048, Longitude: 55.2708},
		DeliveryLocation: Location{Latitude: 25.2070, Longitude: 55.2750},
	}
	repo.reassignCandidates = []*Delivery{repo.deliveries[deliveryID]}
	repo.availableCouriers = []*CourierCandidate{{CourierID: newCourier, Location: Location{Latitude: 25.2050, Longitude: 55.2710}}}

	svc := New(repo, FeatureFlags{
		AutoDispatchEnabled: true,
		AutoReassignEnabled: true,
		DispatchRadiusKM:    20,
		ReassignTimeout:     3 * time.Minute,
		AverageSpeedKPH:     25,
	}, noopLogger{}, NewMockMapsProvider(25), nil, nil)

	reassigned, err := svc.AutoReassignTimedOut(context.Background())
	if err != nil {
		t.Fatalf("AutoReassignTimedOut returned error: %v", err)
	}
	if reassigned != 1 {
		t.Fatalf("expected 1 reassigned delivery, got %d", reassigned)
	}
	if repo.unassignCount != 1 {
		t.Fatalf("expected unassign to be called once, got %d", repo.unassignCount)
	}
	if repo.deliveries[deliveryID].CourierID == nil || *repo.deliveries[deliveryID].CourierID != newCourier {
		t.Fatalf("expected courier to be reassigned")
	}
}

func TestMonitorSLASendsAlerts(t *testing.T) {
	repo := newFakeRepository()
	deliveryID := uuid.New()
	repo.overdueDeliveries = []*Delivery{{ID: deliveryID, Status: "picked_up"}}

	alerts := &alertingSpy{}
	svc := New(repo, FeatureFlags{
		SLAMonitoringEnabled:        true,
		ProviderIntegrationsEnabled: true,
		SLAThreshold:                5 * time.Minute,
	}, noopLogger{}, NewMockMapsProvider(25), alerts, nil)

	count, err := svc.MonitorSLA(context.Background())
	if err != nil {
		t.Fatalf("MonitorSLA returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 SLA alert result, got %d", count)
	}
	if alerts.count != 1 {
		t.Fatalf("expected alert provider to be called once, got %d", alerts.count)
	}
}
