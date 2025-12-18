package routes

import (
	"github.com/gofiber/fiber/v2"

	"uas/app/services"
	"uas/app/repositories"
	"uas/databases"
)

func RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	registerAuthRoutes(api)

	achievementService := services.NewAchievementService(
		repositories.NewAchievementMongoRepository(
			databases.MongoDB,
		),
		repositories.NewAchievementReferenceRepo(
			databases.PSQL,
		),
		repositories.NewStudentRepository(
			databases.PSQL,
		),
		repositories.NewLecturerRepository(
			databases.PSQL,
		),
	)

	registerAchievementRoutes(api, achievementService)

	registerUserRoutes(api)
	registerStudentRoutes(api)
	registerLecturerRoutes(api)
	registerReportRoutes(api)
}
