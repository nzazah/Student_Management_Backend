package services

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"uas/app/models"
	"uas/app/repositories"
	"go.mongodb.org/mongo-driver/bson"
)

type IAchievementService interface {
	Create(c *fiber.Ctx) error
	Submit(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
	List(c *fiber.Ctx) error 
	GetByID(c *fiber.Ctx) error 
	Update(c *fiber.Ctx) error 
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
    ctx := context.Background()
    user := c.Locals("user").(*models.JWTClaims)
    id := c.Params("id")

    // 1️⃣ Ambil student.ID dari user.ID
    student, err := s.studentRepo.FindByUserID(user.UserID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "student profile not found"})
    }

    // 2️⃣ Ambil dokumen achievement dari MongoDB
    ach, err := s.mongoRepo.FindByID(ctx, id)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
    }

    // 3️⃣ Cek apakah achievement milik student ini
    if ach.StudentID != student.ID {
        return c.Status(403).JSON(fiber.Map{"error": "unauthorized"})
    }

    // 4️⃣ Ambil reference Postgres berdasarkan MongoID
    ref, err := s.refRepo.GetByMongoID(id)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "achievement reference not found"})
    }

    // 5️⃣ Pastikan status masih draft
    if ref.Status != "draft" {
        return c.Status(400).JSON(fiber.Map{"error": "only draft can be submitted"})
    }

    // 6️⃣ Update status ke submitted
    now := time.Now()
    if err := s.refRepo.UpdateStatusByMongoID(id, "submitted", &now); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(fiber.Map{"message": "submitted"})
}


//
// FR-005 — Soft Delete Draft
//
func (s *achievementService) Delete(c *fiber.Ctx) error {
    ctx := context.Background()
    user := c.Locals("user").(*models.JWTClaims)
    id := c.Params("id")

    // Ambil student berdasarkan userID
    student, err := s.studentRepo.FindByUserID(user.UserID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "student profile not found"})
    }

    // Ambil reference Postgres berdasarkan MongoID
    ref, err := s.refRepo.GetByMongoID(id)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "achievement reference not found"})
    }

    // Cek authorization berdasarkan student ID
    if ref.StudentID != student.ID {
        return c.Status(403).JSON(fiber.Map{"error": "unauthorized"})
    }

    if ref.Status != "draft" {
        return c.Status(400).JSON(fiber.Map{"error": "only draft can be deleted"})
    }

    // Soft delete di MongoDB
    if err := s.mongoRepo.SoftDelete(ctx, id); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    // Soft delete di Postgres: ubah status menjadi 'deleted'
    if err := s.refRepo.SoftDeleteByMongoID(id); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(fiber.Map{"message": "deleted"})
}


//
// GET /api/v1/achievements
//
func (s *achievementService) List(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.JWTClaims)
	ctx := context.Background()

	var results []fiber.Map

	// =========================
	// ROLE: MAHASISWA
	// =========================
	if user.Role == "student" {

		student, err := s.studentRepo.FindByUserID(user.UserID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "student profile not found"})
		}

		refs, err := s.refRepo.GetByStudentID(student.ID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		for _, ref := range refs {
			doc, err := s.mongoRepo.FindByID(ctx, ref.MongoAchievementID)
			if err != nil {
				continue
			}

			results = append(results, fiber.Map{
				"id":              doc.ID,
				"student_id":      ref.StudentID,
				"achievementType": doc.AchievementType,
				"title":           doc.Title,
				"description":     doc.Description,
				"details":         doc.Details,
				"attachments":     doc.Attachments,
				"tags":            doc.Tags,
				"points":          doc.Points,
				"status":          ref.Status,
				"submittedAt":     ref.SubmittedAt,
				"createdAt":       doc.CreatedAt,
				"updatedAt":       doc.UpdatedAt,
			})
		}

		return c.JSON(results)
	}

	// =========================
	// ROLE: DOSEN WALI (FR-006)
	// =========================
	if user.Role == "lecturer" {

		// 1️⃣ Ambil mahasiswa bimbingan
		students, err := s.studentRepo.FindByAdvisorID(user.UserID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		if len(students) == 0 {
			return c.JSON(results)
		}

		// 2️⃣ Ambil semua prestasi mahasiswa bimbingan
		for _, student := range students {

			refs, err := s.refRepo.GetByStudentID(student.ID)
			if err != nil {
				continue
			}

			for _, ref := range refs {
				doc, err := s.mongoRepo.FindByID(ctx, ref.MongoAchievementID)
				if err != nil {
					continue
				}

				results = append(results, fiber.Map{
					"id":              doc.ID,
					"student_id":      student.ID,
					"achievementType": doc.AchievementType,
					"title":           doc.Title,
					"description":     doc.Description,
					"details":         doc.Details,
					"attachments":     doc.Attachments,
					"tags":            doc.Tags,
					"points":          doc.Points,
					"status":          ref.Status,
					"submittedAt":     ref.SubmittedAt,
					"createdAt":       doc.CreatedAt,
					"updatedAt":       doc.UpdatedAt,
				})
			}
		}

		return c.JSON(results)
	}

	// =========================
	// ROLE: ADMIN
	// =========================
	if user.Role == "admin" {

		docs, err := s.mongoRepo.FindAll(ctx, bson.M{})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		for _, doc := range docs {

			ref, _ := s.refRepo.GetByMongoID(doc.ID)

			var status string
			var submittedAt *time.Time

			if ref != nil {
				status = ref.Status
				submittedAt = ref.SubmittedAt
			}

			results = append(results, fiber.Map{
				"id":              doc.ID,
				"student_id":      ref.StudentID,
				"achievementType": doc.AchievementType,
				"title":           doc.Title,
				"description":     doc.Description,
				"details":         doc.Details,
				"attachments":     doc.Attachments,
				"tags":            doc.Tags,
				"points":          doc.Points,
				"status":          status,
				"submittedAt":     submittedAt,
				"createdAt":       doc.CreatedAt,
				"updatedAt":       doc.UpdatedAt,
			})
		}

		return c.JSON(results)
	}

	return c.Status(403).JSON(fiber.Map{"error": "role not allowed"})
}


//
// GET /api/v1/achievements/:id
//
func (s *achievementService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx := context.Background()

	doc, err := s.mongoRepo.FindByID(ctx, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
	}

	return c.JSON(doc)
}

//
// PUT /api/v1/achievements/:id
//
func (s *achievementService) Update(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.JWTClaims)
	id := c.Params("id")
	ctx := context.Background()

	ref, err := s.refRepo.GetByMongoID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "achievement reference not found"})
	}

	if ref.StudentID != user.UserID {
		return c.Status(403).JSON(fiber.Map{"error": "unauthorized"})
	}

	if ref.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{"error": "only draft can be updated"})
	}

	var payload models.MongoAchievement
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Update fields (title, description, details, attachments)
	err = s.mongoRepo.Update(ctx, id, &payload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "updated"})
}
