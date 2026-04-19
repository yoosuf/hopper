package delivery

import (
	"context"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/yoosuf/hopper/internal/platform/logger"
)

// MockMapsProvider is a local fallback maps implementation.
type MockMapsProvider struct {
	averageSpeedKPH float64
}

// NewMockMapsProvider creates a mock maps provider for route estimates.
func NewMockMapsProvider(averageSpeedKPH float64) *MockMapsProvider {
	if averageSpeedKPH <= 0 {
		averageSpeedKPH = 25
	}
	return &MockMapsProvider{averageSpeedKPH: averageSpeedKPH}
}

// EstimateRoute estimates route distance and ETA with a direct-distance heuristic.
func (m *MockMapsProvider) EstimateRoute(_ context.Context, from, to Location) (*RouteEstimate, error) {
	distanceKM := haversineKM(from, to)
	hours := distanceKM / m.averageSpeedKPH
	etaMinutes := int(math.Ceil(hours * 60))
	if etaMinutes < 1 {
		etaMinutes = 1
	}
	return &RouteEstimate{DistanceKM: distanceKM, ETAMinutes: etaMinutes}, nil
}

// LogAlertingProvider is a logger-backed SLA alerting provider.
type LogAlertingProvider struct {
	log logger.Logger
}

// NewLogAlertingProvider creates a logger-backed alerting provider.
func NewLogAlertingProvider(log logger.Logger) *LogAlertingProvider {
	return &LogAlertingProvider{log: log}
}

// SendSLAAlert emits an alert event to structured logs.
func (p *LogAlertingProvider) SendSLAAlert(_ context.Context, deliveryID uuid.UUID, message string) error {
	p.log.Warn("SLA alert triggered", logger.F("delivery_id", deliveryID), logger.F("message", message))
	return nil
}

// NewMapsProviderFromName returns the configured maps provider with safe fallback.
func NewMapsProviderFromName(name, googleMapsAPIKey, mapboxAPIKey string, averageSpeedKPH float64) MapsProvider {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "google_maps":
		if googleMapsAPIKey != "" {
			return NewGoogleMapsProvider(googleMapsAPIKey)
		}
	case "mapbox":
		if mapboxAPIKey != "" {
			return NewMapboxProvider(mapboxAPIKey)
		}
	}

	return NewMockMapsProvider(averageSpeedKPH)
}
