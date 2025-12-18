package routes

import (
	"github.com/gofiber/fiber/v2"

	"uas/app/services"
	"uas/middleware"
)

func registerStudentRoutes(api fiber.Router) {
	students := api.Group("/students", middleware.JWTProtected())

	students.Get("/", middleware.RequirePermission("student:list"), services.GetStudents)
	students.Get("/:id", middleware.RequirePermission("student:read"), services.GetStudentByID)
	students.Get("/:id/achievements", middleware.RequirePermission("student:achievements"), services.GetStudentAchievements)
	students.Put("/:id/advisor", middleware.RequirePermission("student:update_advisor"), services.UpdateStudentAdvisor)
}
