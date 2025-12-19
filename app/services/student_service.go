package services

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"uas/app/repositories"
	"uas/databases"
)

// GetStudents godoc
// @Summary Get all students
// @Description Mengambil daftar semua mahasiswa
// @Tags Student
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /students [get]
func GetStudents(c *fiber.Ctx) error {
	studentRepo := repositories.NewStudentRepository(databases.PSQL)

	students, err := studentRepo.FindAll()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"data": students})
}

// GetStudentByID godoc
// @Summary Get student by ID
// @Description Mengambil data mahasiswa berdasarkan ID
// @Tags Student
// @Accept json
// @Produce json
// @Param id path string true "Student ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /students/{id} [get]
func GetStudentByID(c *fiber.Ctx) error {
	studentRepo := repositories.NewStudentRepository(databases.PSQL)

	studentID := c.Params("id")
	student, err := studentRepo.FindByID(studentID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "student not found")
	}

	return c.JSON(fiber.Map{"data": student})
}

// GetStudentAchievements godoc
// @Summary Get student's achievements
// @Description Mengambil semua prestasi yang dimiliki mahasiswa tertentu
// @Tags Student
// @Accept json
// @Produce json
// @Param id path string true "Student ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /students/{id}/achievements [get]
func GetStudentAchievements(c *fiber.Ctx) error {
	ctx := context.Background()

	studentRepo := repositories.NewStudentRepository(databases.PSQL)
	refRepo := repositories.NewAchievementReferenceRepo(databases.PSQL)
	mongoRepo := repositories.NewAchievementMongoRepository(databases.MongoDB)

	studentID := c.Params("id")

	student, err := studentRepo.FindByID(studentID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "student not found")
	}

	refs, err := refRepo.GetByStudentID(studentID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var results []fiber.Map

	for _, ref := range refs {
		mongoAch, err := mongoRepo.FindByID(ctx, ref.MongoAchievementID)
		if err != nil {
			continue
		}

		results = append(results, fiber.Map{
			"id":             ref.ID,
			"status":         ref.Status,
			"submitted_at":   ref.SubmittedAt,
			"verified_at":    ref.VerifiedAt,
			"verified_by":    ref.VerifiedBy,
			"rejection_note": ref.RejectionNote,
			"achievement":   mongoAch,
		})
	}

	return c.JSON(fiber.Map{
		"student": student,
		"data":    results,
	})
}

// UpdateStudentAdvisor godoc
// @Summary Update student's advisor
// @Description Mengubah dosen pembimbing mahasiswa tertentu
// @Tags Student
// @Accept json
// @Produce json
// @Param id path string true "Student ID"
// @Param body body map[string]string true "Advisor payload {\"advisor_id\": \"...\"}"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /students/{id}/advisor [put]
func UpdateStudentAdvisor(c *fiber.Ctx) error {
	studentRepo := repositories.NewStudentRepository(databases.PSQL)

	studentID := c.Params("id")

	var payload struct {
		AdvisorID string `json:"advisor_id"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if payload.AdvisorID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "advisor_id is required")
	}

	if _, err := studentRepo.FindByID(studentID); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "student not found")
	}

	if err := studentRepo.UpdateAdvisor(studentID, payload.AdvisorID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"message": "advisor updated successfully",
	})
}
