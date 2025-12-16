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
	UploadAttachment(c *fiber.Ctx) error
	History(c *fiber.Ctx) error
}

type achievementService struct {
    mongoRepo   repositories.IAchievementMongoRepository
    refRepo     repositories.IAchievementReferenceRepo
    studentRepo repositories.IStudentRepository
	lecturerRepo repositories.ILecturerRepository
}


func NewAchievementService(
    mongo repositories.IAchievementMongoRepository,
    ref repositories.IAchievementReferenceRepo,
    student repositories.IStudentRepository,
	lecturer repositories.ILecturerRepository,
) IAchievementService {
    return &achievementService{
        mongoRepo:   mongo,
        refRepo:     ref,
        studentRepo: student,
		lecturerRepo: lecturer,
    }
}

//
// FR-003 — Create Achievement (draft)
//
func (s *achievementService) Create(c *fiber.Ctx) error {
	ctx := context.Background()
	user := c.Locals("user").(*models.JWTClaims)

	// Ambil student dari user
	student, err := s.studentRepo.FindByUserID(user.UserID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "student profile not found",
		})
	}

	var payload models.MongoAchievement
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	now := time.Now()

	// Inject data sistem
	payload.StudentID = student.ID
	payload.Attachments = []models.AchievementAttachment{}
	payload.Points = 0
	payload.CreatedAt = now
	payload.UpdatedAt = now

	// Insert ke MongoDB
	mongoID, err := s.mongoRepo.Insert(ctx, &payload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Simpan reference di PostgreSQL
	ref := &models.AchievementReference{
		ID:                 uuid.New().String(),
		StudentID:          student.ID,
		MongoAchievementID: mongoID,
		Status:             "draft",
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if _, err := s.refRepo.Create(ref); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	payload.ID = mongoID

	return c.Status(201).JSON(payload)
}

func (s *achievementService) UploadAttachment(c *fiber.Ctx) error {
	ctx := context.Background()
	id := c.Params("id")

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "file is required",
		})
	}

	// contoh simpan file (pseudo)
	fileURL := "/uploads/" + file.Filename

	attachment := models.AchievementAttachment{
		FileName:   file.Filename,
		FileUrl:    fileURL,
		FileType:   file.Header.Get("Content-Type"),
		UploadedAt: time.Now(),
	}

	// Push attachment ke MongoDB
	if err := s.mongoRepo.AddAttachment(ctx, id, attachment); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "attachment uploaded",
		"file":    attachment,
	})
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
	if user.Role == "Mahasiswa" {

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
	if user.Role == "Dosen Wali" {

	lecturer, err := s.lecturerRepo.FindByUserID(user.UserID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "lecturer profile not found"})
	}

	students, err := s.studentRepo.FindByAdvisorID(lecturer.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	for _, student := range students {

		refs, err := s.refRepo.GetByStudentID(student.ID)
		if err != nil {
			continue
		}

		for _, ref := range refs {

			// opsional sesuai FR-006
			if ref.Status != "submitted" {
				continue
			}

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
	if user.Role == "Admin" {

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
    ctx := context.Background()
    user := c.Locals("user").(*models.JWTClaims)
    id := c.Params("id")

    var payload struct {
        Points int `json:"points"`
    }

    if err := c.BodyParser(&payload); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "invalid request body",
        })
    }

    if payload.Points <= 0 {
        return c.Status(400).JSON(fiber.Map{
            "error": "points must be greater than 0",
        })
    }

    ref, err := s.refRepo.GetByMongoID(id)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
    }

    if ref.Status != "submitted" {
        return c.Status(400).JSON(fiber.Map{
            "error": "only submitted achievement can be verified",
        })
    }

    // 1️⃣ Update points di MongoDB
    if err := s.mongoRepo.UpdatePoints(ctx, id, payload.Points); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    // 2️⃣ Update status di PostgreSQL
    if err := s.refRepo.VerifyByMongoID(
        id,
        user.UserID,
        time.Now(),
    ); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(fiber.Map{
        "status": "verified",
        "points": payload.Points,
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

func (s *achievementService) History(c *fiber.Ctx) error {
    mongoID := c.Params("id")

    history, err := s.refRepo.GetHistoryByMongoID(mongoID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(fiber.Map{
        "achievement_id": mongoID,
        "history":        history,
    })
}
