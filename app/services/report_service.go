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

func (s *ReportService) GetStudentReport(c *fiber.Ctx) error {
	ctx := context.Background()
	studentID := c.Params("id")
	ids, err := s.ReportRepo.GetVerifiedAchievementMongoIDsByStudent(studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	achievements, err := s.MongoRepo.FindByIDs(ctx, ids)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}


	var totalPoints int
	pointsByType := make(map[string][]int)

	for _, a := range achievements {
		typeName, _ := a["achievementType"].(string)
		if typeName == "" {
			typeName = "Lainnya"
		}

		var pointsVal int
		if val, ok := a["points"]; ok {
			switch v := val.(type) {
			case int32:
				pointsVal = int(v)
			case int64:
				pointsVal = int(v)
			case float64:
				pointsVal = int(v)
			case int:
				pointsVal = v
			}
		}

		totalPoints += pointsVal
		pointsByType[typeName] = append(pointsByType[typeName], pointsVal)
	}

	avgByType := make(map[string]float64)
	for k, arr := range pointsByType {
		sum := 0
		for _, v := range arr {
			sum += v
		}
		if len(arr) > 0 {
			avgByType[k] = float64(sum) / float64(len(arr))
		} else {
			avgByType[k] = 0
		}
	}

	return c.JSON(fiber.Map{
		"student_id":             studentID,
		"total_verified":         len(ids),
		"total_points":           totalPoints,
		"average_points_by_type": avgByType,
		"achievements":           achievements,
	})
}