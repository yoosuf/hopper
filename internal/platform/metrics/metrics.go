package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Service handles metrics collection
type Service struct {
	enabled bool
	port    int

	// Counters
	requestsTotal       *prometheus.CounterVec
	ordersCreated       prometheus.Counter
	deliveriesCompleted prometheus.Counter

	// Gauges
	activeOrders   prometheus.Gauge
	activeCouriers prometheus.Gauge

	// Histograms
	requestDuration *prometheus.HistogramVec
	orderValue      *prometheus.HistogramVec

	mu sync.RWMutex
}

// New creates a new metrics service
func New(enabled bool, port int) *Service {
	s := &Service{
		enabled: enabled,
		port:    port,
	}

	if enabled {
		s.initMetrics()
	}

	return s
}

// initMetrics initializes Prometheus metrics
func (s *Service) initMetrics() {
	s.requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	s.ordersCreated = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "Total number of orders created",
		},
	)

	s.deliveriesCompleted = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "deliveries_completed_total",
			Help: "Total number of deliveries completed",
		},
	)

	s.activeOrders = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_orders",
			Help: "Number of currently active orders",
		},
	)

	s.activeCouriers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_couriers",
			Help: "Number of currently active couriers",
		},
	)

	s.requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	s.orderValue = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_value_cents",
			Help:    "Order value in cents",
			Buckets: []float64{500, 1000, 2000, 5000, 10000, 20000, 50000},
		},
		[]string{"region_id"},
	)

	// Register metrics with default registry
	prometheus.MustRegister(
		s.requestsTotal,
		s.ordersCreated,
		s.deliveriesCompleted,
		s.activeOrders,
		s.activeCouriers,
		s.requestDuration,
		s.orderValue,
	)
}

// Handler returns the HTTP handler for the metrics endpoint
func (s *Service) Handler() http.Handler {
	if !s.enabled {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		})
	}
	return promhttp.Handler()
}

// IncrementCounter increments a counter metric
func (s *Service) IncrementCounter(name string, labels map[string]string) {
	if !s.enabled {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	switch name {
	case "http_requests_total":
		if method, ok := labels["method"]; ok {
			if path, ok := labels["path"]; ok {
				if status, ok := labels["status"]; ok {
					s.requestsTotal.WithLabelValues(method, path, status).Inc()
				}
			}
		}
	case "orders_created_total":
		s.ordersCreated.Inc()
	case "deliveries_completed_total":
		s.deliveriesCompleted.Inc()
	}
}

// RecordDuration records a duration metric
func (s *Service) RecordDuration(name string, duration float64, labels map[string]string) {
	if !s.enabled {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	switch name {
	case "http_request_duration_seconds":
		if method, ok := labels["method"]; ok {
			if path, ok := labels["path"]; ok {
				s.requestDuration.WithLabelValues(method, path).Observe(duration)
			}
		}
	}
}

// RecordGauge records a gauge metric
func (s *Service) RecordGauge(name string, value float64, labels map[string]string) {
	if !s.enabled {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	switch name {
	case "active_orders":
		s.activeOrders.Set(value)
	case "active_couriers":
		s.activeCouriers.Set(value)
	}
}

// ObserveOrderValue records an order value in the histogram
func (s *Service) ObserveOrderValue(regionID string, valueCents float64) {
	if !s.enabled {
		return
	}
	s.orderValue.WithLabelValues(regionID).Observe(valueCents)
}

// IncrementActiveOrders increments the active orders gauge
func (s *Service) IncrementActiveOrders() {
	if !s.enabled {
		return
	}
	s.activeOrders.Inc()
}

// DecrementActiveOrders decrements the active orders gauge
func (s *Service) DecrementActiveOrders() {
	if !s.enabled {
		return
	}
	s.activeOrders.Dec()
}
