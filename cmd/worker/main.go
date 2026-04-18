package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/crewdigital/hopper/internal/platform/config"
	"github.com/crewdigital/hopper/internal/platform/db"
	"github.com/crewdigital/hopper/internal/platform/logger"
	"github.com/crewdigital/hopper/internal/worker"
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
	workerService := worker.New(jobQueue)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker in background
	go workerService.Start(ctx)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down worker...")
	cancel()
	log.Info("Worker exited")
}
