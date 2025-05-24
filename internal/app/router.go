package app

import (
	"net/http"

	"github.com/ElissaDesign/vpn-dev/internal/handler"
	"github.com/go-chi/chi/v5"
)

func SetupRouter(h *handler.UserHandler) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Grouped under /users
	r.Route("/users", func(r chi.Router) {
		r.Get("/", h.GetUsers)
		r.Post("/", h.CreateUser)
	})

	// Later, you can do:
	// r.Mount("/auth", authRoutes())
	// r.Mount("/admin", adminRoutes())

	return r
}
