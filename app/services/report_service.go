package services

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"uas/app/repositories"
)

type ReportService struct {
	reportRepo repositories.ReportRepository
	mongoRepo  repositories.AchievementMongoReportRepository
}

func NewReportService(reportRepo repositories.ReportRepository, mongoRepo repositories.AchievementMongoReportRepository) *ReportService {
	return &ReportService{
		reportRepo: reportRepo,
		mongoRepo:  mongoRepo,
	}
}

func (s *ReportService) Statistics(c *fiber.Ctx) error {
	ctx := context.Background()
	ids, err := s.reportRepo.GetVerifiedAchievementMongoIDs()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	totalPoints, _ := s.mongoRepo.SumPointsByIDs(ctx, ids)
	byType, _ := s.mongoRepo.CountByType(ctx, ids)

	averagePoints := 0.0
	if len(ids) > 0 {
		averagePoints = float64(totalPoints) / float64(len(ids))
	}

	return c.JSON(fiber.Map{
		"total_verified_achievements": len(ids),
		"total_points":                totalPoints,
		"average_points":              averagePoints,
		"achievement_by_type":         byType,
	})
}

func (s *ReportService) StudentReport(c *fiber.Ctx) error {
	ctx := context.Background()
	studentID := c.Params("id")

	ids, err := s.reportRepo.GetVerifiedAchievementMongoIDsByStudent(studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	achievements, _ := s.mongoRepo.FindByIDs(ctx, ids)
	totalPoints, _ := s.mongoRepo.SumPointsByIDs(ctx, ids)

	// Average per type
	pointsByType := make(map[string][]int)
	for _, a := range achievements {
		t := a["achievementType"].(string)
		p := int(a["points"].(int32))
		pointsByType[t] = append(pointsByType[t], p)
	}

	averageByType := make(map[string]float64)
	for k, arr := range pointsByType {
		sum := 0
		for _, v := range arr {
			sum += v
		}
		averageByType[k] = float64(sum) / float64(len(arr))
	}

	return c.JSON(fiber.Map{
		"student_id":            studentID,
		"total_points":          totalPoints,
		"achievements":          achievements,
		"average_points_by_type": averageByType,
	})
}
