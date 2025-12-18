package routes

import (
	"github.com/gofiber/fiber/v2"

	"uas/app/services"
	"uas/middleware"
)

func registerLecturerRoutes(api fiber.Router) {
	lecturers := api.Group("/lecturers", middleware.JWTProtected())

	lecturers.Get("/", middleware.RequirePermission("lecturer:list"), services.GetLecturers)
	lecturers.Get("/:id/advisees", middleware.RequirePermission("lecturer:advisees"), services.GetLecturerAdvisees)
}
