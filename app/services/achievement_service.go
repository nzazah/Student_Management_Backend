package services

import (
	"context"
	"strconv"
	"time"
	"strings"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"uas/app/models"
	"uas/app/repositories"
)

type AchievementService struct {
	MongoRepo    repositories.IAchievementMongoRepository
	RefRepo      repositories.IAchievementReferenceRepo
	StudentRepo  repositories.IStudentRepository
	LecturerRepo repositories.ILecturerRepository
}

func NewAchievementService(
	mongo repositories.IAchievementMongoRepository,
	ref repositories.IAchievementReferenceRepo,
	student repositories.IStudentRepository,
	lecturer repositories.ILecturerRepository,
) *AchievementService {
	return &AchievementService{
		MongoRepo:    mongo,
		RefRepo:      ref,
		StudentRepo:  student,
		LecturerRepo: lecturer,
	}
}

// CreateAchievement godoc
// @Summary Create new achievement
// @Description Mahasiswa menambahkan prestasi baru
// @Tags Achievements
// @Accept json
// @Produce json
// @Param achievement body models.MongoAchievement true "Achievement Data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /achievements [post]
func (s *AchievementService) CreateAchievement() fiber.Handler {
	return func(c *fiber.Ctx) error {

		user := c.Locals("user").(*models.JWTClaims)

		student, err := s.StudentRepo.FindByUserID(user.UserID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "student profile not found"})
		}

		var payload models.MongoAchievement
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		now := time.Now()
		payload.StudentID = student.ID
		payload.Attachments = []models.AchievementAttachment{}
		payload.Points = 0
		payload.CreatedAt = now
		payload.UpdatedAt = now

		mongoID, err := s.MongoRepo.Insert(context.Background(), &payload)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		ref := &models.AchievementReference{
			ID:                 uuid.New().String(),
			StudentID:          student.ID,
			MongoAchievementID: mongoID,
			Status:             "draft",
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		if _, err := s.RefRepo.Create(ref); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(201).JSON(fiber.Map{
			"id":     mongoID,
			"title":  payload.Title,
			"status": "draft",
		})
	}
}

// ListAchievements godoc
// @Summary List achievements
// @Description Melihat daftar prestasi berdasarkan role (mahasiswa, dosen, admin)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Limit per page"
// @Success 200 {array} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /achievements [get]
func (s *AchievementService) ListAchievements() fiber.Handler {
	return func(c *fiber.Ctx) error {

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

		results := []fiber.Map{}

		switch strings.ToLower(user.Role) {

		case "mahasiswa":
			student, err := s.StudentRepo.FindByUserID(user.UserID)
			if err != nil {
				return c.Status(404).JSON(fiber.Map{"error": "student profile not found"})
			}

			refs, _ := s.RefRepo.GetByStudentID(student.ID)
			for _, ref := range refs {
				doc, err := s.MongoRepo.FindByID(ctx, ref.MongoAchievementID)
				if err != nil {
					continue
				}
				results = append(results, fiber.Map{
					"id":     ref.MongoAchievementID,
					"title":  doc.Title,
					"status": ref.Status,
				})
			}

		case "dosen wali", "dosen_wali":
			lecturer, err := s.LecturerRepo.FindByUserID(user.UserID)
			if err != nil {
				return c.Status(404).JSON(fiber.Map{"error": "lecturer profile not found"})
			}

			students, _ := s.StudentRepo.FindByAdvisorID(lecturer.ID)
			for _, student := range students {
				refs, _ := s.RefRepo.GetByStudentID(student.ID)
				for _, ref := range refs {
					if ref.Status != "submitted" {
						continue
					}
					doc, err := s.MongoRepo.FindByID(ctx, ref.MongoAchievementID)
					if err != nil {
						continue
					}
					results = append(results, fiber.Map{
						"id":     ref.MongoAchievementID,
						"title":  doc.Title,
						"status": ref.Status,
					})
				}
			}

		case "admin":
			refs, _ := s.RefRepo.GetAll()
			for _, ref := range refs {
				doc, err := s.MongoRepo.FindByID(ctx, ref.MongoAchievementID)
				if err != nil {
					continue
				}
				results = append(results, fiber.Map{
					"id":        ref.MongoAchievementID,
					"title":     doc.Title,
					"status":    ref.Status,
					"studentId": ref.StudentID,
				})
			}
		}

		end := offset + limit
		if offset > len(results) {
			results = []fiber.Map{}
		} else if end > len(results) {
			results = results[offset:]
		} else {
			results = results[offset:end]
		}

		return c.JSON(results)
	}
}

// GetAchievementByID godoc
// @Summary Get achievement by ID
// @Description Melihat detail prestasi tertentu
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /achievements/{id} [get]
func (s *AchievementService) GetAchievementByID() fiber.Handler {
	return func(c *fiber.Ctx) error {

		id := c.Params("id")
		ctx := context.Background()

		ref, err := s.RefRepo.GetByMongoID(id)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "not found"})
		}

		doc, err := s.MongoRepo.FindByID(ctx, id)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "not found"})
		}

		return c.JSON(fiber.Map{
			"id":          id,
			"title":       doc.Title,
			"description": doc.Description,
			"details":     doc.Details,
			"attachments": doc.Attachments,
			"points":      doc.Points,
			"status":      ref.Status,
		})
	}
}

// UpdateAchievement godoc
// @Summary Update achievement
// @Description Memperbarui data prestasi
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID"
// @Param achievement body models.MongoAchievement true "Achievement Data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /achievements/{id} [put]
func (s *AchievementService) UpdateAchievement() fiber.Handler {
	return func(c *fiber.Ctx) error {

		id := c.Params("id")
		var payload models.MongoAchievement

		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		if err := s.MongoRepo.Update(context.Background(), id, &payload); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"message": "updated"})
	}
}

// DeleteAchievement godoc
// @Summary Delete achievement
// @Description Menghapus prestasi (soft delete)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID"
// @Success 200 {object} map[string]string
// @Security ApiKeyAuth
// @Router /achievements/{id} [delete]
func (s *AchievementService) DeleteAchievement() fiber.Handler {
	return func(c *fiber.Ctx) error {

		id := c.Params("id")

		_ = s.MongoRepo.SoftDelete(context.Background(), id)
		_ = s.RefRepo.SoftDeleteByMongoID(id)

		return c.JSON(fiber.Map{"message": "deleted"})
	}
}


// SubmitAchievement godoc
// @Summary Submit achievement
// @Description Mahasiswa mengirim prestasi untuk diverifikasi
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /achievements/{id}/submit [post]
func (s *AchievementService) SubmitAchievement() fiber.Handler {
	return func(c *fiber.Ctx) error {

		id := c.Params("id")
		now := time.Now()

		if err := s.RefRepo.UpdateStatusByMongoID(id, "submitted", &now); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"message": "submitted"})
	}
}

// VerifyAchievement godoc
// @Summary Verify achievement
// @Description Dosen/admin memverifikasi prestasi
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID"
// @Param points body object true "Points to assign" 
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /achievements/{id}/verify [post]
func (s *AchievementService) VerifyAchievement() fiber.Handler {
	return func(c *fiber.Ctx) error {

		id := c.Params("id")

		var payload struct {
			Points int `json:"points"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		user := c.Locals("user").(*models.JWTClaims)

		if err := s.MongoRepo.UpdatePoints(context.Background(), id, payload.Points); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to update points: " + err.Error()})
		}

		if err := s.RefRepo.VerifyByMongoID(id, user.UserID, time.Now()); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to verify achievement: " + err.Error()})
		}

		return c.JSON(fiber.Map{"status": "verified"})
	}
}

// RejectAchievement godoc
// @Summary Reject achievement
// @Description Dosen/admin menolak prestasi
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID"
// @Param rejection_note body object true "Reason for rejection"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /achievements/{id}/reject [post]
func (s *AchievementService) RejectAchievement() fiber.Handler {
    return func(c *fiber.Ctx) error {

        id := c.Params("id")

        var payload struct {
            RejectionNote string `json:"rejection_note"`
        }

        if err := c.BodyParser(&payload); err != nil {
            return c.Status(400).JSON(fiber.Map{"error": err.Error()})
        }

        // --- GANTI BARIS LAMA DENGAN KODE DI BAWAH INI ---
        if err := s.RefRepo.RejectByMongoID(id, payload.RejectionNote); err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "failed to reject: " + err.Error()})
        }
        // ------------------------------------------------

        return c.JSON(fiber.Map{"status": "rejected"})
    }
}

// GetAchievementHistory godoc
// @Summary Get achievement history
// @Description Melihat riwayat status prestasi
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID"
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /achievements/{id}/history [get]
func (s *AchievementService) GetAchievementHistory() fiber.Handler {
	return func(c *fiber.Ctx) error {

		id := c.Params("id")

		history, err := s.RefRepo.GetHistoryByMongoID(id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(history)
	}
}

// UploadAttachment godoc
// @Summary Upload attachments
// @Description Mahasiswa mengunggah file pendukung prestasi sebelum submit
// @Tags Achievements
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Achievement ID"
// @Param file formData file true "File to upload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /achievements/{id}/attachments [post]
func (s *AchievementService) UploadAttachment() fiber.Handler {
	return func(c *fiber.Ctx) error {

		id := c.Params("id")
		user := c.Locals("user").(*models.JWTClaims)
		ctx := context.Background()

		student, err := s.StudentRepo.FindByUserID(user.UserID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "student profile not found"})
		}

		ref, err := s.RefRepo.GetByMongoID(id)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
		}

		if ref.StudentID != student.ID {
			return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
		}

		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "failed to read form: " + err.Error()})
		}

		files := form.File["files"]
		if len(files) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "no files uploaded"})
		}

		var attachments []models.AchievementAttachment

		for _, file := range files {
			dst := "./uploads/" + file.Filename
			if err := c.SaveFile(file, dst); err != nil {
				continue
			}

			attachments = append(attachments, models.AchievementAttachment{
				FileName:   file.Filename,
				FileUrl:    dst,
				FileType:   file.Header.Get("Content-Type"),
				UploadedAt: time.Now(),
			})
		}

		for _, att := range attachments {
			if err := s.MongoRepo.AddAttachment(ctx, id, att); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "failed to save attachment: " + err.Error()})
			}
		}

		return c.JSON(fiber.Map{
			"message":     "files uploaded",
			"attachments": attachments,
		})
	}
}

