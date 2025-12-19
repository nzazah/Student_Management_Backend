package routes

import (
	"github.com/gofiber/fiber/v2"

	"uas/app/repositories" 
	"uas/app/services"
	"uas/databases"        
	"uas/middleware"
)

func registerAuthRoutes(api fiber.Router) {
	userRepo := repositories.NewUserRepository(databases.PSQL)
	refreshRepo := repositories.NewRefreshRepository(databases.PSQL)

	auth := api.Group("/auth")

	auth.Post("/login", services.Login(userRepo, refreshRepo))
	
	auth.Post("/refresh", services.Refresh)
	auth.Post("/logout", middleware.JWTProtected(), services.Logout)
	auth.Get("/profile", middleware.JWTProtected(), services.Profile)
}