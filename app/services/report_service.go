package services

import (
	"context"
	"uas/app/repositories"

	"github.com/gofiber/fiber/v2"
)

type ReportService struct {
    ReportRepo repositories.IReportRepository
    MongoRepo  repositories.IAchievementMongoReportRepository 
}

func NewReportService(r repositories.IReportRepository, m repositories.IAchievementMongoReportRepository) *ReportService {
    return &ReportService{
        ReportRepo: r,
        MongoRepo:  m,
    }
}

// GetAchievementStatistics godoc
// @Summary Get achievement statistics
// @Description Mendapatkan statistik prestasi mahasiswa (total, rata-rata, dan per tipe)
// @Tags Report
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /reports/statistics [get]
func (s *ReportService) GetAchievementStatistics(c *fiber.Ctx) error {
	ctx := context.Background()

	ids, err := s.ReportRepo.GetVerifiedAchievementMongoIDs()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	totalPoints, _ := s.MongoRepo.SumPointsByIDs(ctx, ids)
	byType, _ := s.MongoRepo.CountByType(ctx, ids)

	avg := 0.0
	if len(ids) > 0 {
		avg = float64(totalPoints) / float64(len(ids))
	}

	return c.JSON(fiber.Map{
		"total_verified_achievements": len(ids),
		"total_points":                totalPoints,
		"average_points":               avg,
		"achievement_by_type":         byType,
	})
}

// GetStudentReport godoc
// @Summary Get student achievement report
// @Description Mendapatkan laporan prestasi mahasiswa tertentu beserta poin dan rata-rata per tipe
// @Tags Report
// @Accept json
// @Produce json
// @Param id path string true "Student ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /reports/students/{id} [get]
func (s *ReportService) GetStudentReport(c *fiber.Ctx) error {
	ctx := context.Background()
	studentID := c.Params("id")

	ids, err := s.ReportRepo.GetVerifiedAchievementMongoIDsByStudent(studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	achievements, _ := s.MongoRepo.FindByIDs(ctx, ids)
	totalPoints, _ := s.MongoRepo.SumPointsByIDs(ctx, ids)

	pointsByType := make(map[string][]int)
	for _, a := range achievements {
		typeName, okT := a["achievementType"].(string)
		
		var pointsVal int
		if p, ok := a["points"].(int32); ok {
			pointsVal = int(p)
		} else if p, ok := a["points"].(int64); ok {
			pointsVal = int(p)
		} else if p, ok := a["points"].(int); ok {
			pointsVal = p
		}

		if okT {
			pointsByType[typeName] = append(pointsByType[typeName], pointsVal)
		}
	}

	avgByType := make(map[string]float64)
	for k, arr := range pointsByType {
		sum := 0
		for _, v := range arr {
			sum += v
		}
		avgByType[k] = float64(sum) / float64(len(arr))
	}

	return c.JSON(fiber.Map{
		"student_id":             studentID,
		"total_points":           totalPoints,
		"achievements":           achievements,
		"average_points_by_type": avgByType,
	})
}