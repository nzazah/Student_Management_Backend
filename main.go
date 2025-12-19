// @title UAS Achievement API
// @version 1.0
// @description API untuk mengelola prestasi mahasiswa
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @host localhost:3000
// @BasePath /api/v1
// @schemes http

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	_ "uas/docs"
	"uas/config"
	"uas/databases"
	"uas/routes"
)

func main() {
	config.LoadEnv()

	databases.ConnectPostgres()
	databases.ConnectMongoDB()

	app := fiber.New()
	
	app.Get("/swagger/*", swagger.HandlerDefault) // default: http://localhost:3000/swagger/index.html

	app.Use(cors.New(cors.Config{
        AllowOrigins: "*",
        AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
    }))

	routes.RegisterRoutes(app)

	log.Println("Server running at http://localhost:3000")
	log.Fatal(app.Listen(":3000"))
}
