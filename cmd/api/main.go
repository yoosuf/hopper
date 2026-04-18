package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	_ "github.com/crewdigital/hopper/docs"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/crewdigital/hopper/internal/admin"
	"github.com/crewdigital/hopper/internal/audit"
	"github.com/crewdigital/hopper/internal/auth"
	"github.com/crewdigital/hopper/internal/delivery"
	"github.com/crewdigital/hopper/internal/jobs"
	"github.com/crewdigital/hopper/internal/menus"
	"github.com/crewdigital/hopper/internal/notifications"
	"github.com/crewdigital/hopper/internal/orders"
	"github.com/crewdigital/hopper/internal/payments"
	"github.com/crewdigital/hopper/internal/platform/config"
	"github.com/crewdigital/hopper/internal/platform/db"
	"github.com/crewdigital/hopper/internal/platform/httpx"
	"github.com/crewdigital/hopper/internal/platform/idempotency"
	"github.com/crewdigital/hopper/internal/platform/logger"
	"github.com/crewdigital/hopper/internal/platform/metrics"
	"github.com/crewdigital/hopper/internal/platform/middleware"
	"github.com/crewdigital/hopper/internal/platform/validator"
	"github.com/crewdigital/hopper/internal/promotions"
	"github.com/crewdigital/hopper/internal/regions"
	"github.com/crewdigital/hopper/internal/restaurants"
	"github.com/crewdigital/hopper/internal/reviews"
	"github.com/crewdigital/hopper/internal/support"
	"github.com/crewdigital/hopper/internal/tax"
	"github.com/crewdigital/hopper/internal/users"
)

// @title           Uber Eats Clone API
// @version         1.0
// @description     This is a sample server for an Uber Eats clone API. You can visit the Swagger UI documentation at /swagger/index.html
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Initialize logger
	log := logger.New(cfg)
	log.Info("Starting Hopper Food Delivery API")

	// Initialize database
	dbPool, err := db.New(cfg)
	if err != nil {
		log.Error("Failed to connect to database", logger.F("error", err))
		panic(err)
	}
	defer dbPool.Close()

	// Initialize validator
	validator := validator.New()

	// Initialize metrics
	metrics := metrics.New(cfg.Metrics.Enabled, cfg.Metrics.Port)

	// Initialize job queue
	jobQueue := make(chan interface{}, 100)

	// Initialize repositories
	authRepo := auth.NewRepository(dbPool.Pool)
	userRepo := users.NewRepository(dbPool.Pool)
	regionRepo := regions.NewRepository(dbPool.Pool)
	taxRepo := tax.NewRepository(dbPool.Pool)
	restaurantRepo := restaurants.NewRepository(dbPool.Pool)
	menuRepo := menus.NewRepository(dbPool.Pool)
	orderRepo := orders.NewRepository(dbPool.Pool)
	deliveryRepo := delivery.NewRepository(dbPool.Pool)
	paymentRepo := payments.NewRepository(dbPool.Pool)
	notificationRepo := notifications.NewRepository(dbPool.Pool)
	reviewRepo := reviews.NewRepository(dbPool.Pool)
	promotionRepo := promotions.NewRepository(dbPool.Pool)
	supportRepo := support.NewRepository(dbPool.Pool)
	adminRepo := admin.NewRepository(dbPool.Pool)
	auditRepo := audit.NewRepository(dbPool.Pool)
	idempotencyRepo := idempotency.NewRepository(dbPool.Pool)

	// Initialize services
	authService := auth.New(authRepo, cfg.JWT.Secret, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	userService := users.New(userRepo)
	regionService := regions.New(regionRepo)
	taxService := tax.New(taxRepo)
	restaurantService := restaurants.New(restaurantRepo)
	menuService := menus.New(menuRepo)
	orderService := orders.New(orderRepo)
	deliveryService := delivery.New(deliveryRepo)
	paymentService := payments.New(paymentRepo)
	notificationService := notifications.New(notificationRepo, nil)
	reviewService := reviews.New(reviewRepo)
	promotionService := promotions.New(promotionRepo)
	supportService := support.New(supportRepo)
	_ = admin.New(adminRepo)
	_ = jobs.New(jobQueue)
	_ = audit.New(auditRepo)
	_ = idempotency.New(idempotencyRepo, 24*time.Hour)

	// Initialize handlers
	authHandler := auth.NewHandler(authService, validator)
	userHandler := users.NewHandler(userService, validator)
	regionHandler := regions.NewHandler(regionService)
	taxHandler := tax.NewHandler(taxService, validator)
	supportHandler := support.NewHandler(supportService, validator)
	restaurantHandler := restaurants.NewHandler(restaurantService, validator)
	menuHandler := menus.NewHandler(menuService, validator)
	orderHandler := orders.NewHandler(orderService, validator)
	deliveryHandler := delivery.NewHandler(deliveryService)
	paymentHandler := payments.NewHandler(paymentService, validator)
	notificationHandler := notifications.NewHandler(notificationService)
	reviewHandler := reviews.NewHandler(reviewService, validator)
	promotionHandler := promotions.NewHandler(promotionService, validator)

	// Setup router
	router := chi.NewRouter()

	// CORS middleware
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Logging middleware
	router.Use(middleware.Logging(log))

	// Recovery middleware
	router.Use(middleware.RequestID)

	// Public routes
	router.Post("/auth/register", authHandler.Register)
	router.Post("/auth/login", authHandler.Login)

	// Protected routes
	router.Group(func(r chi.Router) {
		r.Use(middleware.Auth(authService))

		// User routes
		r.Get("/users/me", userHandler.GetProfile)
		r.Put("/users/me", userHandler.UpdateProfile)

		// Region routes
		r.Get("/regions", regionHandler.ListRegions)
		r.Get("/regions/{id}", regionHandler.GetRegion)

		// Tax routes
		r.Get("/tax/categories", taxHandler.ListTaxCategories)
		r.Get("/tax/zones", taxHandler.ListTaxZones)
		r.Get("/tax/rates", taxHandler.ListTaxRates)

		// Restaurant routes
		r.Get("/restaurants", restaurantHandler.ListRestaurants)
		r.Get("/restaurants/{id}", restaurantHandler.GetRestaurant)
		r.Post("/restaurants", restaurantHandler.CreateRestaurant)
		r.Put("/restaurants/{id}", restaurantHandler.UpdateRestaurant)
		r.Get("/restaurants/{id}/hours", restaurantHandler.GetRestaurantHours)
		r.Put("/restaurants/{id}/hours", restaurantHandler.SetRestaurantHours)

		// Menu routes
		r.Post("/menus", menuHandler.CreateMenuItem)
		r.Put("/menus/{id}", menuHandler.UpdateMenuItem)
		r.Delete("/menus/{id}", menuHandler.DeleteMenuItem)

		// Order routes
		r.Get("/orders/{id}", orderHandler.GetOrder)
		r.Post("/orders", orderHandler.CreateOrder)
		r.Put("/orders/{id}/cancel", orderHandler.CancelOrder)

		// Delivery routes
		r.Get("/deliveries/{id}", deliveryHandler.GetDelivery)
		r.Put("/deliveries/{id}/status", deliveryHandler.UpdateDeliveryStatus)

		// Payment routes
		r.Get("/payments/{id}", paymentHandler.GetPayment)
		r.Post("/payments", paymentHandler.CreatePayment)
		r.Put("/payments/{id}/status", paymentHandler.UpdatePaymentStatus)

		// Notification routes
		r.Get("/notifications/{id}", notificationHandler.GetNotification)
		r.Put("/notifications/{id}/read", notificationHandler.MarkAsRead)
		r.Put("/notifications/read-all", notificationHandler.MarkAllAsRead)

		// Support routes
		r.Post("/support/tickets", supportHandler.CreateTicket)
		r.Get("/support/tickets", supportHandler.ListUserTickets)
		r.Get("/support/tickets/{id}", supportHandler.GetTicket)
		r.Put("/support/tickets/{id}", supportHandler.UpdateTicket)
		r.Post("/support/tickets/{id}/messages", supportHandler.CreateMessage)
		r.Get("/support/tickets/{id}/messages", supportHandler.ListMessages)

		// Review routes
		r.Post("/reviews", reviewHandler.CreateReview)
		r.Get("/reviews/{id}", reviewHandler.GetReview)
		r.Put("/reviews/{id}", reviewHandler.UpdateReview)
		r.Delete("/reviews/{id}", reviewHandler.DeleteReview)
		r.Get("/reviews/my", reviewHandler.ListUserReviews)
		r.Get("/reviews/{target_type}/{target_id}", reviewHandler.ListTargetReviews)
		r.Get("/reviews/{target_type}/{target_id}/stats", reviewHandler.GetTargetRatingStats)

		// Promotion routes
		r.Post("/promotions", promotionHandler.CreatePromotion)
		r.Get("/promotions", promotionHandler.ListPromotions)
		r.Get("/promotions/{id}", promotionHandler.GetPromotion)
		r.Get("/promotions/code/{code}", promotionHandler.GetPromotionByCode)
		r.Put("/promotions/{id}", promotionHandler.UpdatePromotion)
		r.Delete("/promotions/{id}", promotionHandler.DeletePromotion)
		r.Post("/promotions/validate", promotionHandler.ValidatePromotion)
		r.Post("/promotions/use", promotionHandler.UsePromotion)
	})

	// Swagger documentation routes
	router.Get("/swagger/*", httpSwagger.WrapHandler)
	router.Get("/swagger/doc.json", httpSwagger.WrapHandler)

	// Health check
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteSuccess(w, http.StatusOK, map[string]string{
			"status": "healthy",
		})
	})

	// Metrics endpoint
	router.Handle("/metrics", metrics.Handler())

	// Start server
	server := &http.Server{
		Addr:         ":" + string(rune(cfg.Server.Port)),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Info("Server starting on port " + string(rune(cfg.Server.Port)))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed to start", logger.F("error", err))
			panic(err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", logger.F("error", err))
		panic(err)
	}

	log.Info("Server exited")
}
