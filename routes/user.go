package routes

import (
	"github.com/gofiber/fiber/v2"

	"uas/app/services"
	"uas/app/repositories"
	"uas/databases"
	"uas/middleware"
)

func RegisterUserRoutes(api fiber.Router) {
	users := api.Group(
		"/users",
		middleware.JWTProtected(),
		middleware.RequirePermission("user:manage"),
	)

	userRepo := repositories.NewUserRepository(databases.PSQL)
	studentRepo := repositories.NewStudentRepository(databases.PSQL)
	lecturerRepo := repositories.NewLecturerRepository(databases.PSQL)

	users.Get("/", services.GetAllUsers)
	users.Get("/:id", services.GetUserByID)
	users.Post("/", services.CreateUser(userRepo, studentRepo, lecturerRepo))
	users.Put("/:id", services.UpdateUser)
	users.Delete("/:id", services.DeleteUser)
	users.Put("/:id/role", services.AssignRole)
}
