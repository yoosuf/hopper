package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yoosuf/hopper/internal/delivery"
	"github.com/yoosuf/hopper/internal/platform/config"
	"github.com/yoosuf/hopper/internal/platform/db"
	"github.com/yoosuf/hopper/internal/platform/logger"
	"github.com/yoosuf/hopper/internal/worker"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Initialize logger
	log := logger.New(cfg)
	log.Info("Starting Hopper Food Delivery Worker")

	// Initialize database
	dbPool, err := db.New(cfg)
	if err != nil {
		log.Error("Failed to connect to database", logger.F("error", err))
		panic(err)
	}
	defer dbPool.Close()

	// Initialize job queue
	jobQueue := make(chan worker.Job, 100)

	// Initialize worker service
	workerService := worker.New(jobQueue, log, cfg.Worker.Concurrency)
	deliveryRepo := delivery.NewRepository(dbPool.Pool)
	deliveryFlags := delivery.FeatureFlags{
		AutoDispatchEnabled:         cfg.Courier.AutoDispatchEnabled,
		RouteOptimizationEnabled:    cfg.Courier.RouteOptimizationEnabled,
		LiveTrackingEnabled:         cfg.Courier.LiveTrackingEnabled,
		AutoReassignEnabled:         cfg.Courier.AutoReassignEnabled,
		SLAMonitoringEnabled:        cfg.Courier.SLAMonitoringEnabled,
		ProviderIntegrationsEnabled: cfg.Courier.ProviderIntegrationsEnabled,
		DispatchRadiusKM:            cfg.Courier.DispatchRadiusKM,
		ReassignTimeout:             cfg.Courier.ReassignTimeout,
		SLAThreshold:                cfg.Courier.SLAThreshold,
		AverageSpeedKPH:             cfg.Courier.AverageSpeedKPH,
	}
	deliveryService := delivery.New(
		deliveryRepo,
		deliveryFlags,
		log,
		delivery.NewMapsProviderFromName(
			cfg.Courier.MapsProvider,
			cfg.Courier.GoogleMapsAPIKey,
			cfg.Courier.MapboxAPIKey,
			cfg.Courier.AverageSpeedKPH,
		),
		delivery.NewLogAlertingProvider(log),
		nil,
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker in background
	go workerService.Start(ctx)

	// Run courier automation loop in worker process.
	go func() {
		ticker := time.NewTicker(cfg.Worker.PollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				reassigned, err := deliveryService.AutoReassignTimedOut(ctx)
				if err != nil {
					log.Warn("Courier auto-reassign loop failed", logger.F("error", err))
				} else if reassigned > 0 {
					log.Info("Courier auto-reassign completed", logger.F("reassigned_count", reassigned))
				}

				alerts, err := deliveryService.MonitorSLA(ctx)
				if err != nil {
					log.Warn("Courier SLA monitor loop failed", logger.F("error", err))
				} else if alerts > 0 {
					log.Info("Courier SLA monitor completed", logger.F("alerts", alerts))
				}
			}
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down worker...")
	cancel()
	log.Info("Worker exited")
}
