package db

import (
	"context"
	"fmt"
	"time"

	"github.com/crewdigital/hopper/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps pgxpool.Pool for database operations
type Pool struct {
	*pgxpool.Pool
}

// New creates a new database connection pool
func New(cfg *config.Config) (*Pool, error) {
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.Database.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.Database.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.Database.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = time.Minute * 5
	poolConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Pool{Pool: pool}, nil
}

// Close closes the database connection pool
func (p *Pool) Close() {
	p.Pool.Close()
}

// Health checks if the database connection is healthy
func (p *Pool) Health(ctx context.Context) error {
	return p.Ping(ctx)
}
