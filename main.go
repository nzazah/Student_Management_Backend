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
	// Load .env
	config.LoadEnv()

	// Connect to Postgres
	database.ConnectPostgres()

	// Init Fiber
	app := fiber.New()

	// Repository & Service
	userRepo := repositories.NewUserRepository(database.DB)
	authService := services.NewAuthService(userRepo)

	// Setup routes
	routes.Setup(app, authService)

	// Informasi server
	log.Println("Server running at http://localhost:3000")

	// Start server
	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
