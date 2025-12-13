package services

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"uas/app/models"
	"uas/app/repositories"
	"strconv"
)

type IAchievementService interface {
	Create(c *fiber.Ctx) error
	Submit(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
	List(c *fiber.Ctx) error 
	GetByID(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Verify(c *fiber.Ctx) error
	Reject(c *fiber.Ctx) error 
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

	// =========================
	// Pagination
	// =========================
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

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
	}

	// =========================
	// ROLE: DOSEN WALI (FR-006)
	// =========================
	if user.Role == "lecturer" {

		students, err := s.studentRepo.FindByAdvisorID(user.UserID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

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
	}

	// =========================
	// ROLE: ADMIN (FR-010)
	// =========================
	if user.Role == "admin" {

		refs, err := s.refRepo.GetAll()
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
	}

	// =========================
	// Pagination FINAL
	// =========================
	end := offset + limit
	if offset > len(results) {
		results = []fiber.Map{}
	} else {
		if end > len(results) {
			end = len(results)
		}
		results = results[offset:end]
	}

	return c.JSON(results)
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

func (s *achievementService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx := context.Background()

	// Ambil reference dari PostgreSQL
	ref, err := s.refRepo.GetByMongoID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "achievement reference not found",
		})
	}

	// Ambil detail dari MongoDB
	doc, err := s.mongoRepo.FindByID(ctx, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "achievement not found",
		})
	}

	return c.JSON(fiber.Map{
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

func (s *achievementService) Verify(c *fiber.Ctx) error {
    user := c.Locals("user").(*models.JWTClaims)
    id := c.Params("id")

    // 1️⃣ Ambil reference
    ref, err := s.refRepo.GetByMongoID(id)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
    }

    // 2️⃣ Validasi status
    if ref.Status != "submitted" {
        return c.Status(400).JSON(fiber.Map{
            "error": "only submitted achievement can be verified",
        })
    }

    // 3️⃣ Verify
    err = s.refRepo.VerifyByMongoID(
        id,
        user.UserID,     // verified_by
        time.Now(),      // verified_at
    )
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(fiber.Map{
        "status": "verified",
    })
}

func (s *achievementService) Reject(c *fiber.Ctx) error {
    id := c.Params("id")

    var payload struct {
        RejectionNote string `json:"rejection_note"`
    }

    if err := c.BodyParser(&payload); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

    if payload.RejectionNote == "" {
        return c.Status(400).JSON(fiber.Map{
            "error": "rejection_note is required",
        })
    }

    ref, err := s.refRepo.GetByMongoID(id)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
    }

    if ref.Status != "submitted" {
        return c.Status(400).JSON(fiber.Map{
            "error": "only submitted achievement can be rejected",
        })
    }

    err = s.refRepo.RejectByMongoID(id, payload.RejectionNote)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(fiber.Map{
        "status": "rejected",
    })
}
