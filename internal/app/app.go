package app

import (
	"log"
	"net/http"

	"github.com/ElissaDesign/vpn-dev/internal/config"
	"github.com/ElissaDesign/vpn-dev/internal/handler"
	"github.com/ElissaDesign/vpn-dev/internal/repository"
	"github.com/ElissaDesign/vpn-dev/internal/service"
)

func Start() {
	cfg := config.Load()
	db := config.ConnectMongo(cfg)
	// Initialize your repository (pass DB connection or config as needed)
	userRepo := repository.NewUserRepository(db)

	// Initialize your service with the repository
	userService := service.NewUserService(userRepo)

	// Initialize your handler with the service
	userHandler := handler.NewUserHandler(userService)

	r := SetupRouter(userHandler)

	log.Println("🌐 Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
