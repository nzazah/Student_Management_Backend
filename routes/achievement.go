package routes

import (
	"github.com/gofiber/fiber/v2"

	"uas/app/services"
	"uas/middleware"
)

func registerAchievementRoutes(
	api fiber.Router,
	achService *services.AchievementService,
) {

	ach := api.Group(
		"/achievements",
		middleware.JWTProtected(),
	)

	ach.Get(
		"/",
		middleware.RequirePermission("achievement:list"),
		achService.ListAchievements(),
	)

	ach.Get(
		"/:id",
		middleware.RequirePermission("achievement:view"),
		achService.GetAchievementByID(),
	)

	ach.Post(
		"/",
		middleware.RequirePermission("achievement:create"),
		achService.CreateAchievement(),
	)

	ach.Put(
		"/:id",
		middleware.RequirePermission("achievement:update"),
		achService.UpdateAchievement(),
	)

	ach.Delete(
		"/:id",
		middleware.RequirePermission("achievement:delete"),
		achService.DeleteAchievement(),
	)

	ach.Post(
		"/:id/submit",
		middleware.RequirePermission("achievement:submit"),
		achService.SubmitAchievement(),
	)

	ach.Post(
		"/:id/verify",
		middleware.RequirePermission("achievement:verify"),
		achService.VerifyAchievement(),
	)

	ach.Post(
		"/:id/reject",
		middleware.RequirePermission("achievement:reject"),
		achService.RejectAchievement(),
	)

	ach.Get(
		"/:id/history",
		middleware.RequirePermission("achievement:view"),
		achService.GetAchievementHistory(),
	)

	ach.Post(
	"/:id/attachments",
	middleware.RequirePermission("achievement:upload_attachment"),
	achService.UploadAttachment(),
)

}

