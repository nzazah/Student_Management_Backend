package services_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"uas/app/mocks"
	"uas/app/models"
	"uas/app/services"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginUnit(t *testing.T) {
	app := fiber.New()
	userMock := new(mocks.UserRepoMock)
	refreshMock := new(mocks.RefreshRepoMock)

	app.Post("/login", services.Login(userMock, refreshMock))

	passwordRaw := "password123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(passwordRaw), 10)

	t.Run("Success Login", func(t *testing.T) {
		dummyUser := &models.User{
			ID: "user-1", Username: "admin", PasswordHash: string(hashed), IsActive: true,
		}

		userMock.On("FindByUsername", mock.Anything, "admin").Return(dummyUser, nil).Once()
		userMock.On("GetRoleName", mock.Anything, mock.Anything).Return("Admin", nil).Once()
		userMock.On("GetPermissionsByUserID", mock.Anything, "user-1").Return([]string{"read"}, nil).Once()
		refreshMock.On("Delete", "user-1").Return(nil).Once()
		refreshMock.On("Save", "user-1", mock.Anything).Return(nil).Once()

		payload := map[string]string{"username": "admin", "password": passwordRaw}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Invalid Password", func(t *testing.T) {
		dummyUser := &models.User{Username: "admin", PasswordHash: string(hashed), IsActive: true}
		userMock.On("FindByUsername", mock.Anything, "admin").Return(dummyUser, nil).Once()

		payload := map[string]string{"username": "admin", "password": "wrongpassword"}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)
		assert.Equal(t, 401, resp.StatusCode)
	})
}