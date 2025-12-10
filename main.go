package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"uas/config"
	"uas/databases"
	"uas/app/repositories"
	"uas/app/services"
	"uas/routes"
)

func main() {
	config.LoadEnv()

	db := databases.ConnectPostgres()

	app := fiber.New()

	userRepo := repositories.NewUserRepository(db)
	studentRepo := repositories.NewStudentRepository(db)
	lecturerRepo := repositories.NewLecturerRepository(db)
	refreshRepo := repositories.NewRefreshRepository(db)


	authService := services.NewAuthService(
	userRepo,
	studentRepo,
	lecturerRepo,
	refreshRepo,
	)

	routes.Setup(app, authService, userRepo)


	log.Println("Server running at http://localhost:3000")

	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
