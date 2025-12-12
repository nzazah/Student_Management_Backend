package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"

	"uas/config"
	"uas/databases"
	"uas/app/repositories"
	"uas/app/services"
	"uas/routes"
)

func main() {
	config.LoadEnv()

	// CONNECT POSTGRES
	pg, err := databases.ConnectPostgres()
	if err != nil {
		log.Fatal("Failed to connect PostgreSQL:", err)
	}
	defer pg.Close()

	// CONNECT MONGODB
	mongoUri := os.Getenv("MONGO_URI")
	mongo, err := databases.ConnectMongo(mongoUri)
	if err != nil {
		log.Fatal("Failed to connect MongoDB:", err)
	}
	defer mongo.Disconnect(context.Background())

	app := fiber.New()

	// ============= REPOSITORIES =============

	userRepo := repositories.NewUserRepository(pg)
	studentRepo := repositories.NewStudentRepository(pg)
	lecturerRepo := repositories.NewLecturerRepository(pg)
	refreshRepo := repositories.NewRefreshRepository(pg)

	achievementMongoRepo := repositories.NewAchievementMongoRepository(mongo)
achievementRefRepo := repositories.NewAchievementReferenceRepo(pg)

	// ============= SERVICES =============

	authService := services.NewAuthService(
		userRepo,
		studentRepo,
		lecturerRepo,
		refreshRepo,
	)

	achievementService := services.NewAchievementService(
	achievementMongoRepo,
	achievementRefRepo,
	studentRepo,
	)

	routes.Setup(app, authService, userRepo, achievementService)

	log.Println("Server running at http://localhost:3000")

	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
