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
	// Load environment variable (.env)
	config.LoadEnv()

	// Connect to PostgreSQL
	database.ConnectPostgres()

	// Create Fiber instance
	app := fiber.New()

	// Initialize Repository
	userRepo := repositories.NewUserRepository(database.DB)

	// Initialize Service (service menerima interface)
	authService := services.NewAuthService(userRepo)

	// Setup Routes (tanpa logic di route)
	routes.Setup(app, authService)

	// Run server
	log.Println("🚀 Server running on :8080")
	app.Listen(":8080")
}
