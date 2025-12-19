package services_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"uas/app/mocks"
	"uas/app/services"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
)

func TestReportService(t *testing.T) {
	reportMock := new(mocks.ReportRepoMock)
	mongoMock := new(mocks.MongoReportMock)
	service := services.NewReportService(reportMock, mongoMock)

	app := fiber.New()
	app.Get("/reports/statistics", service.GetAchievementStatistics)
	app.Get("/reports/students/:id", service.GetStudentReport)

	t.Run("GetAchievementStatistics - Success", func(t *testing.T) {
		mockIDs := []string{"id1", "id2"}
		
		reportMock.On("GetVerifiedAchievementMongoIDs").Return(mockIDs, nil).Once()
		mongoMock.On("SumPointsByIDs", mock.Anything, mockIDs).Return(150, nil).Once()
		mongoMock.On("CountByType", mock.Anything, mockIDs).Return(map[string]int{"Lomba": 2}, nil).Once()

		req := httptest.NewRequest("GET", "/reports/statistics", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assert.Equal(t, 75.0, result["average_points"])
		assert.Equal(t, float64(150), result["total_points"])
		
		reportMock.AssertExpectations(t)
		mongoMock.AssertExpectations(t)
	})

	t.Run("GetStudentReport - Success Calculation", func(t *testing.T) {
		studentID := "MHS001"
		mockIDs := []string{"id_a", "id_b"}
		
		reportMock.On("GetVerifiedAchievementMongoIDsByStudent", studentID).Return(mockIDs, nil).Once()
		
		mockData := []bson.M{
			{"achievementType": "Akademik", "points": int32(100)},
			{"achievementType": "Akademik", "points": int32(50)},
		}
		
		mongoMock.On("FindByIDs", mock.Anything, mockIDs).Return(mockData, nil).Once()
		mongoMock.On("SumPointsByIDs", mock.Anything, mockIDs).Return(150, nil).Once()

		req := httptest.NewRequest("GET", "/reports/students/"+studentID, nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		
		avgByType := result["average_points_by_type"].(map[string]interface{})
		assert.Equal(t, 75.0, avgByType["Akademik"])
	})
}