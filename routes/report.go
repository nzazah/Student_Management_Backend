package routes

import (
    "github.com/gofiber/fiber/v2"
    "uas/app/services"
    "uas/middleware"
)

func registerReportRoutes(api fiber.Router, s *services.ReportService) {
    reports := api.Group(
        "/reports",
        middleware.JWTProtected(),
    )

    reports.Get("/statistics", middleware.RequirePermission("report:view"), s.GetAchievementStatistics)
    reports.Get("/student/:id", middleware.RequirePermission("report:view"), s.GetStudentReport)
}