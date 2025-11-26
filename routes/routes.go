package routes

import (
	"github.com/gofiber/fiber/v2"
	"uas/app/services"
)

func Setup(app *fiber.App, authService *services.AuthService) {

	api := app.Group("/api")

	// Hanya memanggil service langsung, TANPA logic di route
	api.Post("/auth/login", authService.Login)
}
