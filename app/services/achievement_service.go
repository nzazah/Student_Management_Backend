package services

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"uas/app/models"
	"uas/app/repositories"
)

type IAchievementService interface {
	Create(c *fiber.Ctx) error
	Submit(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}

type achievementService struct {
    mongoRepo   repositories.IAchievementMongoRepository
    refRepo     repositories.IAchievementReferenceRepo
    studentRepo repositories.IStudentRepository
}


func NewAchievementService(
    mongo repositories.IAchievementMongoRepository,
    ref repositories.IAchievementReferenceRepo,
    student repositories.IStudentRepository,
) IAchievementService {
    return &achievementService{
        mongoRepo:   mongo,
        refRepo:     ref,
        studentRepo: student,
    }
}


//
// FR-003 — Create Achievement (draft)
//
func (s *achievementService) Create(c *fiber.Ctx) error {
    ctx := context.Background()
    user := c.Locals("user").(*models.JWTClaims)

    // GET student ID from user_id
    student, err := s.studentRepo.FindByUserID(user.UserID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{
            "error": "student profile not found",
        })
    }

    var payload models.MongoAchievement
    if err := c.BodyParser(&payload); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

    // Correct student_id
    payload.StudentID = student.ID

    mongoID, err := s.mongoRepo.Insert(ctx, &payload)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    ref := &models.AchievementReference{
        ID:                 uuid.New().String(),
        StudentID:          student.ID,               // FIX HERE
        MongoAchievementID: mongoID,
        Status:             "draft",
        CreatedAt:          time.Now(),
        UpdatedAt:          time.Now(),
    }

    if _, err = s.refRepo.Create(ref); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(payload)
}


//
// FR-004 — Submit Draft (draft → submitted)
//
func (s *achievementService) Submit(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.JWTClaims)

	id := c.Params("id")
	ref, err := s.refRepo.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
	}

	if ref.StudentID != user.UserID {
		return c.Status(403).JSON(fiber.Map{"error": "unauthorized"})
	}

	if ref.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{"error": "only draft can be submitted"})
	}

	now := time.Now()
	if err := s.refRepo.UpdateStatus(id, "submitted", &now); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "submitted"})
}

//
// FR-005 — Soft Delete Draft
//
func (s *achievementService) Delete(c *fiber.Ctx) error {
	ctx := context.Background() // REQUIRED for Mongo
	user := c.Locals("user").(*models.JWTClaims)

	id := c.Params("id")
	ref, err := s.refRepo.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
	}

	if ref.StudentID != user.UserID {
		return c.Status(403).JSON(fiber.Map{"error": "unauthorized"})
	}

	if ref.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{"error": "only draft can be deleted"})
	}

	// Soft delete Mongo data
	if err := s.mongoRepo.SoftDelete(ctx, ref.MongoAchievementID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Soft delete reference in Postgres
	if err := s.refRepo.SoftDelete(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "deleted"})
}
