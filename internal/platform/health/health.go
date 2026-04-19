package health

import (
	"encoding/json"
	"net/http"
)

// Status represents the health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// Check represents a health check result
type Check struct {
	Name    string `json:"name"`
	Status  Status `json:"status"`
	Message string `json:"message,omitempty"`
}

// Response represents the health check response
type Response struct {
	Status  Status  `json:"status"`
	Checks  []Check `json:"checks,omitempty"`
	Message string  `json:"message,omitempty"`
}

// Checker defines the interface for health check implementations
type Checker interface {
	Check() Check
}

// Handler returns an HTTP handler for health checks
func Handler(checkers ...Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := Response{
			Status: StatusHealthy,
			Checks: make([]Check, 0, len(checkers)),
		}

		for _, checker := range checkers {
			check := checker.Check()
			response.Checks = append(response.Checks, check)

			if check.Status == StatusUnhealthy {
				response.Status = StatusUnhealthy
			} else if check.Status == StatusDegraded && response.Status != StatusUnhealthy {
				response.Status = StatusDegraded
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if response.Status == StatusHealthy {
			w.WriteHeader(http.StatusOK)
		} else if response.Status == StatusDegraded {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		json.NewEncoder(w).Encode(response)
	}
}
