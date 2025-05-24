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

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	r := SetupRouter(userHandler)

	log.Println("🌐 Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
