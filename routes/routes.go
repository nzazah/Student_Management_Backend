package routes

import (
	"github.com/gofiber/fiber/v2"
	"uas/app/services"
	"uas/middleware"
)

func Setup(app *fiber.App, auth *services.AuthService) {

	api := app.Group("/api/v1")

	authRoute := api.Group("/auth")

	// Public
	authRoute.Post("/login", auth.Login)
	authRoute.Post("/refresh", auth.Refresh)

	// Protected
	authRoute.Post("/logout", middleware.JWTProtected(), auth.Logout)
	authRoute.Get("/profile", middleware.JWTProtected(), auth.Profile)
}
