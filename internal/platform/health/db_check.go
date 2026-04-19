package health

import (
	"context"
	"time"

	"github.com/yoosuf/hopper/internal/platform/db"
)

// DBChecker checks database health
type DBChecker struct {
	pool *db.Pool
}

// NewDBChecker creates a new database health checker
func NewDBChecker(pool *db.Pool) *DBChecker {
	return &DBChecker{pool: pool}
}

// Check performs the database health check
func (c *DBChecker) Check() Check {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := c.pool.Ping(ctx)
	if err != nil {
		return Check{
			Name:    "database",
			Status:  StatusUnhealthy,
			Message: err.Error(),
		}
	}

	// Get pool metrics to ensure connections are available
	stats := c.pool.Stat()
	if stats.TotalConns() == 0 {
		return Check{
			Name:    "database",
			Status:  StatusDegraded,
			Message: "no active connections",
		}
	}

	return Check{
		Name:   "database",
		Status: StatusHealthy,
	}
}
