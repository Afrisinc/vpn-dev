package app

import (
	"net/http"

	"github.com/ElissaDesign/vpn-dev/internal/config"
	"github.com/ElissaDesign/vpn-dev/internal/handler"
	"github.com/ElissaDesign/vpn-dev/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/ElissaDesign/vpn-dev/docs"
)

// SetupRouter configures the HTTP router with all routes and middleware
func SetupRouter(userHandler *handler.UserHandler, serverHandler *handler.ServerHandler, deviceHandler *handler.DeviceHandler, cfg config.Config, logger *zerolog.Logger) http.Handler {
	r := chi.NewRouter()

	// Global middleware (applied to all routes)
	r.Use(middleware.PanicRecovery)
	r.Use(middleware.RequestID)
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.LimitBodySize(cfg.Security.MaxRequestSize))
	r.Use(middleware.CORS)

	// Health check endpoint (no rate limiting)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Swagger UI - API documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	// Create rate limiter
	rateLimiter := middleware.NewRateLimiter(
		float64(cfg.Security.RateLimitRequests)/60.0, // Convert per-minute to per-second
		cfg.Security.RateLimitRequests,               // Burst size
	)

	// User routes with rate limiting
	r.Route("/users", func(r chi.Router) {
		r.Use(rateLimiter.Middleware)

		r.Post("/register", userHandler.CreateUser)
		r.Post("/toggle", userHandler.ToggleUserHandler)
		r.Get("/", userHandler.GetUsers)
		r.Get("/by-public-key", userHandler.GetUserByPublicKey)
	})

	// Admin routes for server management (no rate limiting on admin)
	r.Route("/admin/servers", func(r chi.Router) {
		r.Get("/", serverHandler.GetAllServers)
		r.Get("/healthy", serverHandler.GetHealthyServers)
		r.Post("/", serverHandler.CreateServer)
		r.Put("/{serverId}/status", serverHandler.UpdateServerStatus)
	})

	// Device and usage routes (with rate limiting from /users)
	r.Route("/users/{userId}/devices", func(r chi.Router) {
		r.Use(rateLimiter.Middleware)

		r.Post("/", deviceHandler.RegisterDevice)
		r.Get("/", deviceHandler.GetUserDevices)
		r.Route("/{deviceId}", func(r chi.Router) {
			r.Get("/", deviceHandler.GetDeviceByID)
			r.Delete("/", deviceHandler.DeleteDevice)
			r.Get("/usage", deviceHandler.GetDeviceUsage)
			r.Get("/config", deviceHandler.GetDeviceConfig)
		})
	})

	// User usage routes (with rate limiting from /users)
	r.Route("/users/{userId}/usage", func(r chi.Router) {
		r.Use(rateLimiter.Middleware)
		r.Get("/", deviceHandler.GetUserUsage)
	})

	return r
}
