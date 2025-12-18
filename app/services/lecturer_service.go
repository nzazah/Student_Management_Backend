package services

import (
	"github.com/gofiber/fiber/v2"
	"uas/app/models" 
	"uas/app/repositories"
	"uas/databases"
)

// GetLecturers godoc
// @Summary List all lecturers
// @Description Mendapatkan daftar semua dosen
// @Tags Lecturer
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /lecturers [get]
func GetLecturers(c *fiber.Ctx) error {
	lecturerRepo := repositories.NewLecturerRepository(databases.PSQL)

	lecturers, err := lecturerRepo.FindAll()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"data": lecturers,
	})
}

// GetLecturerAdvisees godoc
// @Summary Get lecturer's advisees
// @Description Mendapatkan daftar mahasiswa bimbingan dari dosen tertentu
// @Tags Lecturer
// @Accept json
// @Produce json
// @Param id path string true "Lecturer ID"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /lecturers/{id}/advisees [get]
func GetLecturerAdvisees(c *fiber.Ctx) error {
	lecturerID := c.Params("id")

	lecturerRepo := repositories.NewLecturerRepository(databases.PSQL)
	studentRepo := repositories.NewStudentRepository(databases.PSQL)

	lecturer, err := lecturerRepo.FindByID(lecturerID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "lecturer not found")
	}

	claims := c.Locals("user").(*models.JWTClaims)

	if claims.Role == "Dosen" && lecturer.UserID != claims.UserID {
		return fiber.ErrForbidden
	}

	students, err := studentRepo.FindByAdvisorID(lecturerID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"lecturer": lecturer,
		"advisees": students,
	})
}
