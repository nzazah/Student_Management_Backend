package mocks

import "github.com/stretchr/testify/mock"

type ReportRepoMock struct {
	mock.Mock
}

func (m *ReportRepoMock) GetVerifiedAchievementMongoIDs() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *ReportRepoMock) GetVerifiedAchievementMongoIDsByStudent(studentID string) ([]string, error) {
	args := m.Called(studentID)
	return args.Get(0).([]string), args.Error(1)
}