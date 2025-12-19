package mocks

import (
	"time"
	"uas/app/models"
	"github.com/stretchr/testify/mock"
)

type AchievementRefMock struct {
	mock.Mock
}

func (m *AchievementRefMock) UpdateStatusByMongoID(mongoID string, status string, submittedAt *time.Time) error {
	args := m.Called(mongoID, status, submittedAt)
	return args.Error(0)
}


func (m *AchievementRefMock) Create(ref *models.AchievementReference) (string, error) {
	args := m.Called(ref)
	return args.String(0), args.Error(1)
}

func (m *AchievementRefMock) GetByID(id string) (*models.AchievementReference, error) {
	args := m.Called(id)
	return args.Get(0).(*models.AchievementReference), args.Error(1)
}

func (m *AchievementRefMock) SoftDeleteByMongoID(mongoID string) error {
	return m.Called(mongoID).Error(0)
}

func (m *AchievementRefMock) GetByMongoID(mongoID string) (*models.AchievementReference, error) {
	args := m.Called(mongoID)
	return args.Get(0).(*models.AchievementReference), args.Error(1)
}

func (m *AchievementRefMock) GetByStudentID(studentID string) ([]*models.AchievementReference, error) {
	args := m.Called(studentID)
	return args.Get(0).([]*models.AchievementReference), args.Error(1)
}

func (m *AchievementRefMock) GetByStudentIDs(studentIDs []string) ([]*models.AchievementReference, error) {
	args := m.Called(studentIDs)
	return args.Get(0).([]*models.AchievementReference), args.Error(1)
}

func (m *AchievementRefMock) GetAll() ([]*models.AchievementReference, error) {
	args := m.Called()
	return args.Get(0).([]*models.AchievementReference), args.Error(1)
}

func (m *AchievementRefMock) VerifyByMongoID(mongoID string, vBy string, vAt time.Time) error {
	return m.Called(mongoID, vBy, vAt).Error(0)
}

func (m *AchievementRefMock) RejectByMongoID(mongoID string, note string) error {
	return m.Called(mongoID, note).Error(0)
}

func (m *AchievementRefMock) GetHistoryByMongoID(mongoID string) ([]*models.AchievementReference, error) {
	args := m.Called(mongoID)
	return args.Get(0).([]*models.AchievementReference), args.Error(1)
}