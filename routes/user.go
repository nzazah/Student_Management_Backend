package routes

import (
	"github.com/gofiber/fiber/v2"

	"uas/app/services"
	"uas/middleware"
)

func registerUserRoutes(api fiber.Router) {
	users := api.Group(
		"/users",
		middleware.JWTProtected(),
		middleware.RequirePermission("user:manage"),
	)

	users.Get("/", services.GetAllUsers)
	users.Get("/:id", services.GetUserByID)
	users.Post("/", services.CreateUser)
	users.Put("/:id", services.UpdateUser)
	users.Delete("/:id", services.DeleteUser)
	users.Put("/:id/role", services.AssignRole)
}
