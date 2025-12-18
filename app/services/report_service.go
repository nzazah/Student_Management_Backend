package services

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"uas/app/repositories"
	"uas/databases"
)

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
func GetAchievementStatistics(c *fiber.Ctx) error {
	ctx := context.Background()

	reportRepo := repositories.NewReportRepository(databases.PSQL)
	mongoRepo := repositories.NewAchievementMongoReportRepository(databases.MongoDB)

	ids, err := reportRepo.GetVerifiedAchievementMongoIDs()
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	totalPoints, _ := mongoRepo.SumPointsByIDs(ctx, ids)
	byType, _ := mongoRepo.CountByType(ctx, ids)

	avg := 0.0
	if len(ids) > 0 {
		avg = float64(totalPoints) / float64(len(ids))
	}

	return c.JSON(fiber.Map{
		"total_verified_achievements": len(ids),
		"total_points":                totalPoints,
		"average_points":              avg,
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
func GetStudentReport(c *fiber.Ctx) error {
	ctx := context.Background()
	studentID := c.Params("id")

	reportRepo := repositories.NewReportRepository(databases.PSQL)
	mongoRepo := repositories.NewAchievementMongoReportRepository(databases.MongoDB)

	ids, err := reportRepo.GetVerifiedAchievementMongoIDsByStudent(studentID)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	achievements, _ := mongoRepo.FindByIDs(ctx, ids)
	totalPoints, _ := mongoRepo.SumPointsByIDs(ctx, ids)

	pointsByType := make(map[string][]int)
	for _, a := range achievements {
		t := a["achievementType"].(string)
		p := int(a["points"].(int32))
		pointsByType[t] = append(pointsByType[t], p)
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
