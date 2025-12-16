package routes

import (
	"github.com/gofiber/fiber/v2"
	"uas/app/repositories"
	"uas/app/services"
	"uas/middleware"
)

func Setup(
	app *fiber.App,
	auth *services.AuthService,
	userRepo repositories.IUserRepository,
	userMgmtRepo repositories.IUserManagementRepository,
	studentRepo repositories.IStudentRepository,
	lecturerRepo repositories.ILecturerRepository,
	achievementService services.IAchievementService,
) {
	api := app.Group("/api/v1")
	
	authRoute := api.Group("/auth")
	authRoute.Post("/login", auth.Login)
	authRoute.Post("/refresh", auth.Refresh)
	authRoute.Post("/logout", middleware.JWTProtected(), auth.Logout)
	authRoute.Get("/profile", middleware.JWTProtected(), auth.Profile)

	ach := api.Group("/achievements", middleware.JWTProtected())
	ach.Get("/",  middleware.RequirePermission("achievement:list", userRepo), achievementService.List,)
	ach.Get("/:id",  middleware.RequirePermission("achievement:view", userRepo),  achievementService.GetByID,)
	ach.Post("/", middleware.RequirePermission("achievement:create", userRepo), achievementService.Create)
	ach.Put("/:id", middleware.RequirePermission("achievement:update", userRepo), achievementService.Update)
	ach.Delete("/:id", middleware.RequirePermission("achievement:delete", userRepo), achievementService.Delete)
	ach.Post("/:id/submit", middleware.RequirePermission("achievement:submit", userRepo), achievementService.Submit)
	ach.Post("/:id/verify", middleware.JWTProtected(), middleware.RequirePermission("achievement:verify", userRepo), achievementService.Verify,)
	ach.Post("/:id/reject", middleware.JWTProtected(), middleware.RequirePermission("achievement:reject", userRepo), achievementService.Reject,)

	userService := services.NewUserService(
		userMgmtRepo,
		studentRepo,
		lecturerRepo,
	)

	users := api.Group("/users", middleware.JWTProtected(),	middleware.RequirePermission("user:manage", userRepo),)
	users.Get("/", userService.GetAllUsers)
	users.Get("/:id", userService.GetUserByID)
	users.Post("/", userService.CreateUser)
	users.Put("/:id", userService.UpdateUser)
	users.Delete("/:id", userService.DeleteUser)
	users.Put("/:id/role", userService.AssignRole)

	studentService := services.NewStudentService(studentRepo)

	students := api.Group("/students")
	students.Get("/", middleware.JWTProtected(), middleware.RequirePermission("student:list", userRepo), studentService.GetAll,)
	students.Get("/:id", middleware.JWTProtected(), middleware.RequirePermission("student:read", userRepo), studentService.GetByID,)
	students.Get("/:id/achievements", middleware.JWTProtected(), middleware.RequirePermission("student:achievements", userRepo),studentService.GetAchievements,)
	students.Put("/:id/advisor", middleware.JWTProtected(), middleware.RequirePermission("student:update_advisor", userRepo), studentService.UpdateAdvisor,)

}
