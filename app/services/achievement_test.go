package services_test

import (
    "bytes"
    "encoding/json"
    "net/http/httptest"
    "testing"
    "uas/app/mocks"
    "uas/app/models"
    "uas/app/services"
	"errors"

    "github.com/gofiber/fiber/v2"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestAchievementService(t *testing.T) {
    mongoMock := new(mocks.AchievementMongoMock)
    refMock := new(mocks.AchievementRefMock)
    service := &services.AchievementService{
        MongoRepo: mongoMock,
        RefRepo:   refMock,
    }

    app := fiber.New()

    app.Use(func(c *fiber.Ctx) error {
        c.Locals("user", &models.JWTClaims{UserID: "admin-123"})
        return c.Next()
    })

    app.Post("/achievement/:id/submit", service.SubmitAchievement())
    app.Post("/achievement/:id/verify", service.VerifyAchievement())
    app.Post("/achievement/:id/reject", service.RejectAchievement())
    t.Run("Submit - Success", func(t *testing.T) {
        id := "123"
        refMock.On("UpdateStatusByMongoID", id, "submitted", mock.Anything).Return(nil).Once()
        req := httptest.NewRequest("POST", "/achievement/"+id+"/submit", nil)
        resp, _ := app.Test(req)
        assert.Equal(t, 200, resp.StatusCode)
    })

    t.Run("Verify - Success", func(t *testing.T) {
        id := "65818e69d9f58c42a0a6d001"
        payload := map[string]int{"points": 100}
        body, _ := json.Marshal(payload)
        mongoMock.On("UpdatePoints", mock.Anything, id, 100).Return(nil).Once()
        refMock.On("VerifyByMongoID", id, "admin-123", mock.Anything).Return(nil).Once()
        
        req := httptest.NewRequest("POST", "/achievement/"+id+"/verify", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        resp, _ := app.Test(req)
        assert.Equal(t, 200, resp.StatusCode)
    })

    t.Run("Reject - Success", func(t *testing.T) {
        id := "mongo-id-123"
        note := "Berkas tidak lengkap"
        payload := map[string]string{"rejection_note": note}
        body, _ := json.Marshal(payload)

        refMock.On("RejectByMongoID", id, note).Return(nil).Once()

        req := httptest.NewRequest("POST", "/achievement/"+id+"/reject", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        resp, _ := app.Test(req)

        assert.Equal(t, 200, resp.StatusCode)
    })

    t.Run("Reject - Database Error", func(t *testing.T) {
        id := "mongo-id-456"
        note := "Rejected due to error"
        payload := map[string]string{"rejection_note": note}
        body, _ := json.Marshal(payload)

        refMock.On("RejectByMongoID", id, note).
            Return(errors.New("database connection lost")).Once()

        req := httptest.NewRequest("POST", "/achievement/"+id+"/reject", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        resp, _ := app.Test(req)

        assert.Equal(t, 500, resp.StatusCode)
    })
}