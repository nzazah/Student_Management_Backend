package routes

import (
	"github.com/gofiber/fiber/v2"

	"uas/app/services"
	"uas/middleware"
)

func registerAuthRoutes(api fiber.Router) {
	auth := api.Group("/auth")

	auth.Post("/login", services.Login)
	auth.Post("/refresh", services.Refresh)
	auth.Post("/logout", middleware.JWTProtected(), services.Logout)
	auth.Get("/profile", middleware.JWTProtected(), services.Profile)
}
