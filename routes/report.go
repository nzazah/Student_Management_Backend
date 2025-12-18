package routes

import (
	"github.com/gofiber/fiber/v2"

	"uas/app/services"
	"uas/middleware"
)

func registerReportRoutes(api fiber.Router) {
	reports := api.Group(
		"/reports",
		middleware.JWTProtected(),
	)

	reports.Get("/statistics", middleware.RequirePermission("report:view"), services.GetAchievementStatistics)
	reports.Get("/student/:id", middleware.RequirePermission("report:view"), services.GetStudentReport)
}
