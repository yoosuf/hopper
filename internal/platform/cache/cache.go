package cache

import (
	"context"
	"time"

	"github.com/yoosuf/hopper/internal/platform/config"
	"github.com/yoosuf/hopper/internal/platform/logger"
)

// Cache defines the interface for caching operations
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
}

// New creates a new cache instance based on configuration
func New(cfg *config.Config, log logger.Logger) (Cache, error) {
	if cfg.Redis.Enabled {
		// TODO: Implement Redis cache when package is installed
		// For now, fall back to memory cache
		log.Warn("Redis enabled but not available, using memory cache", logger.F("address", cfg.Redis.Address))
	}

	log.Info("Using in-memory cache")
	return NewMemoryCache(log), nil
}
