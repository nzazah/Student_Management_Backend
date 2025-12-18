// @title UAS Achievement API
// @version 1.0
// @description API untuk mengelola prestasi mahasiswa
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @host localhost:3000
// @BasePath /api/v1
// @schemes http

package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/arsmn/fiber-swagger/v2" // benar


	_ "uas/docs" // generated swagger docs
	"uas/config"
	"uas/databases"
	"uas/routes"
)

func main() {
	config.LoadEnv()

	databases.ConnectPostgres()
	databases.ConnectMongoDB()

	app := fiber.New()

	// Swagger route
	app.Get("/swagger/*", swagger.HandlerDefault) // default: http://localhost:3000/swagger/index.html

	routes.RegisterRoutes(app)

	log.Println("Server running at http://localhost:3000")
	log.Fatal(app.Listen(":3000"))
}
