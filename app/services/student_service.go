package services

import (
	"errors"
	"context"
	"github.com/gofiber/fiber/v2"
	"uas/app/models"
	"uas/app/repositories"
)

type StudentService struct {
	studentRepo repositories.IStudentRepository
	achievementRefRepo repositories.IAchievementReferenceRepo
	mongoAchRepo       repositories.IAchievementMongoRepository
}

func NewStudentService(
	studentRepo repositories.IStudentRepository,
	achievementRefRepo repositories.IAchievementReferenceRepo,
	mongoAchRepo repositories.IAchievementMongoRepository,
) *StudentService {
	return &StudentService{
		studentRepo:        studentRepo,
		achievementRefRepo: achievementRefRepo,
		mongoAchRepo:       mongoAchRepo,
	}
}

/* =========================
   GET /students
   (Admin only)
========================= */
func (s *StudentService) GetAll(c *fiber.Ctx) error {
	students, err := s.studentRepo.FindAll()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"data": students,
	})
}

/* =========================
   GET /students/:id
   Admin | Dosen | Mahasiswa (self)
========================= */
func (s *StudentService) GetByID(c *fiber.Ctx) error {
	studentID := c.Params("id")

	student, err := s.studentRepo.FindByID(studentID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "student not found")
	}

	// OPTIONAL: pembatasan mahasiswa
	role := c.Locals("role")
	userID := c.Locals("user_id")

	if role == "student" && student.UserID != userID {
		return fiber.ErrForbidden
	}

	return c.JSON(fiber.Map{
		"data": student,
	})
}

/* =========================
   GET /students/:id/achievements
   Admin | Dosen | Mahasiswa (self)
========================= */
func (s *StudentService) GetAchievements(c *fiber.Ctx) error {
	studentID := c.Params("id")

	student, err := s.studentRepo.FindByID(studentID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "student not found")
	}

	// Restrict Mahasiswa (hanya boleh lihat miliknya)
	role := c.Locals("role")
	userID := c.Locals("user_id")

	if role == "Mahasiswa" && student.UserID != userID {
		return fiber.ErrForbidden
	}

	// 1️⃣ Ambil reference dari PostgreSQL
	refs, err := s.achievementRefRepo.GetByStudentID(studentID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	ctx := context.Background()
	var results []fiber.Map

	// 2️⃣ Ambil detail dari MongoDB
	for _, ref := range refs {
		mongoAch, err := s.mongoAchRepo.FindByID(ctx, ref.MongoAchievementID)
		if err != nil {
			continue // skip kalau data mongo hilang
		}

		results = append(results, fiber.Map{
			"id":              ref.ID,
			"status":          ref.Status,
			"submitted_at":    ref.SubmittedAt,
			"verified_at":     ref.VerifiedAt,
			"verified_by":     ref.VerifiedBy,
			"rejection_note":  ref.RejectionNote,

			// Mongo data
			"achievement": fiber.Map{
				"id":               mongoAch.ID,
				"achievement_type": mongoAch.AchievementType,
				"title":            mongoAch.Title,
				"description":      mongoAch.Description,
				"details":          mongoAch.Details,
				"attachments":      mongoAch.Attachments,
				"tags":             mongoAch.Tags,
				"points":           mongoAch.Points,
				"created_at":       mongoAch.CreatedAt,
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": results,
	})
}


/* =========================
   PUT /students/:id/advisor
   Admin | Dosen
========================= */
func (s *StudentService) UpdateAdvisor(c *fiber.Ctx) error {
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

	_, err := s.studentRepo.FindByID(studentID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "student not found")
	}

	if err := s.studentRepo.UpdateAdvisor(studentID, payload.AdvisorID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"message": "advisor updated successfully",
	})
}

/* =========================
   INTERNAL USE
   (create student)
========================= */
func (s *StudentService) Create(student *models.Student) error {
	if student.UserID == "" {
		return errors.New("user_id is required")
	}
	return s.studentRepo.Create(student)
}

/* =========================
   GET advisees (for lecturer)
========================= */
func (s *StudentService) GetByAdvisorID(c *fiber.Ctx) error {
	advisorID := c.Params("id")

	students, err := s.studentRepo.FindByAdvisorID(advisorID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"data": students,
	})
}

/* =========================
   GET student by user_id
   (for profile / auth)
========================= */
func (s *StudentService) GetByUserID(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "student not found")
	}

	return c.JSON(fiber.Map{
		"data": student,
	})
}
