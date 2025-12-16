package services

import (
	"github.com/gofiber/fiber/v2"
	"uas/app/repositories"
)

type LecturerService struct {
	lecturerRepo repositories.ILecturerRepository
	studentRepo  repositories.IStudentRepository
}

func NewLecturerService(
	lecturerRepo repositories.ILecturerRepository,
	studentRepo repositories.IStudentRepository,
) *LecturerService {
	return &LecturerService{
		lecturerRepo: lecturerRepo,
		studentRepo:  studentRepo,
	}
}

/* =========================
   GET /lecturers
   Admin only
========================= */
func (s *LecturerService) GetAll(c *fiber.Ctx) error {
	lecturers, err := s.lecturerRepo.FindAll()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"data": lecturers,
	})
}

/* =========================
   GET /lecturers/:id/advisees
   Admin | Dosen (self)
========================= */
func (s *LecturerService) GetAdvisees(c *fiber.Ctx) error {
	lecturerID := c.Params("id")

	// cek dosen
	lecturer, err := s.lecturerRepo.FindByID(lecturerID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "lecturer not found")
	}

	// Restrict Dosen: hanya boleh lihat advisee sendiri
	role := c.Locals("role")
	userID := c.Locals("user_id")

	if role == "Dosen" && lecturer.UserID != userID {
		return fiber.ErrForbidden
	}

	students, err := s.studentRepo.FindByAdvisorID(lecturerID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"lecturer": lecturer,
		"advisees": students,
	})
}
