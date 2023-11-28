package routes

import (
	"go-chat-application/handlers"
	"go-chat-application/middleware"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(r chi.Router) {
	r.Get("/readiness", handlers.ReadinessHandler)
	r.Get("/err", handlers.ErrorHandler)
}

func SetupApiRoutes(r chi.Router) {
	r.Post("/users/login", middleware.WithDB(handlers.LoginUserHandler))
	r.Post("/users/create", middleware.WithDB(handlers.CreateUserHandler))
	r.Get("/users", middleware.WithDB(handlers.GetUsersHandler))
	r.Put("/users", middleware.WithDB(handlers.UpdateUserHandler))
	r.Delete("/users", middleware.WithDB(handlers.DeleteUserHandler))
}
