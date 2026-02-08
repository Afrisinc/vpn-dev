package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ElissaDesign/vpn-dev/internal/config"
	"github.com/ElissaDesign/vpn-dev/internal/db"
	userhandler "github.com/ElissaDesign/vpn-dev/internal/handler"
	"github.com/ElissaDesign/vpn-dev/internal/repository"
	"github.com/ElissaDesign/vpn-dev/internal/service"
	"github.com/ElissaDesign/vpn-dev/internal/validator"
	"github.com/ElissaDesign/vpn-dev/pkg/logger"
)

// Start initializes and runs the application
func Start() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.InitLogger(cfg.IsDevelopment)
	appLogger := logger.GetLogger()

	appLogger.Info().Msg("Starting VPN service...")

	// Connect to PostgreSQL
	pool, err := config.ConnectDatabase(cfg)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.ClosePool()

	// Initialize repositories
	userRepo := repository.NewUserRepository(pool)
	serverRepo := repository.NewServerRepository(pool)
	deviceRepo := repository.NewDeviceRepository(pool)
	deviceUsageRepo := repository.NewDeviceUsageRepository(pool)

	// Initialize services
	userService := service.NewUserService(userRepo)
	deviceService := service.NewDeviceService(deviceRepo, deviceUsageRepo, userRepo, serverRepo, appLogger)

	// Initialize validator
	val := validator.New()

	// Initialize handlers
	userHandler := userhandler.NewUserHandler(userService, serverRepo, deviceRepo, deviceUsageRepo, &cfg.WireGuard, val, appLogger)
	serverHandler := userhandler.NewServerHandler(serverRepo, val, appLogger)
	deviceHandler := userhandler.NewDeviceHandler(deviceService, val, appLogger)

	// Setup router with middleware
	r := SetupRouter(userHandler, serverHandler, deviceHandler, *cfg, appLogger)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		appLogger.Info().Msgf("🌐 Server running on :%d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal().Err(err).Msg("Server error")
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info().Msg("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Error().Err(err).Msg("Server forced to shutdown")
		os.Exit(1)
	}

	appLogger.Info().Msg("Server shutdown complete")
}
