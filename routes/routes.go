package routes

import (
	"github.com/gofiber/fiber/v2"

	"uas/app/repositories"
	"uas/app/services"
	"uas/databases"
)

func RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	registerAuthRoutes(api)

	achievementService := services.NewAchievementService(
		repositories.NewAchievementMongoRepository(databases.MongoDB),
		repositories.NewAchievementReferenceRepo(databases.PSQL),
		repositories.NewStudentRepository(databases.PSQL),
		repositories.NewLecturerRepository(databases.PSQL),
	)
	registerAchievementRoutes(api, achievementService)

	reportRepo := repositories.NewReportRepository(databases.PSQL)
	mongoReportRepo := repositories.NewAchievementMongoReportRepository(databases.MongoDB)
	reportService := services.NewReportService(reportRepo, mongoReportRepo)

	registerReportRoutes(api, reportService)

	RegisterUserRoutes(api)
	registerStudentRoutes(api)
	registerLecturerRoutes(api)
}