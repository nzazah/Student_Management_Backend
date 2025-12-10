package routes

import (
	"github.com/gofiber/fiber/v2"
	"uas/app/repositories"
	"uas/app/services"
	"uas/middleware"
)

func Setup(app *fiber.App, auth *services.AuthService, userRepo repositories.IUserRepository) {
	api := app.Group("/api/v1")

	authRoute := api.Group("/auth")
	authRoute.Post("/login", auth.Login)
	authRoute.Post("/refresh", auth.Refresh)
	authRoute.Post("/logout", middleware.JWTProtected(), auth.Logout)
	authRoute.Get("/profile", middleware.JWTProtected(), auth.Profile)

	ach := api.Group("/achievements")
	ach.Post("/", middleware.JWTProtected(), middleware.RequirePermission("achievement:create", userRepo))
	ach.Put("/:id", middleware.JWTProtected(), middleware.RequirePermission("achievement:update", userRepo))
	ach.Post("/:id/verify", middleware.JWTProtected(), middleware.RequirePermission("achievement:verify", userRepo))
}
