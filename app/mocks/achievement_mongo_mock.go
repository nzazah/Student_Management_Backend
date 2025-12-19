package mocks

import (
	"context"
	"uas/app/models"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
)

type AchievementMongoMock struct {
	mock.Mock
}

func (m *AchievementMongoMock) UpdatePoints(ctx context.Context, id string, points int) error {
	args := m.Called(ctx, id, points)
	return args.Error(0)
}

func (m *AchievementMongoMock) Insert(ctx context.Context, data *models.MongoAchievement) (string, error) {
	args := m.Called(ctx, data)
	return args.String(0), args.Error(1)
}

func (m *AchievementMongoMock) FindByID(ctx context.Context, id string) (*models.MongoAchievement, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.MongoAchievement), args.Error(1)
}

func (m *AchievementMongoMock) FindAll(ctx context.Context, filter bson.M) ([]*models.MongoAchievement, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.MongoAchievement), args.Error(1)
}

func (m *AchievementMongoMock) Update(ctx context.Context, id string, data *models.MongoAchievement) error {
	return m.Called(ctx, id, data).Error(0)
}

func (m *AchievementMongoMock) SoftDelete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *AchievementMongoMock) AddAttachment(ctx context.Context, id string, att models.AchievementAttachment) error {
	return m.Called(ctx, id, att).Error(0)
}

func (m *AchievementMongoMock) FindByIDs(ctx context.Context, ids []string) ([]map[string]interface{}, error) {
    args := m.Called(ctx, ids)
    
    var results []map[string]interface{}
    if args.Get(0) != nil {
        results = args.Get(0).([]map[string]interface{})
    }
    
    return results, args.Error(1)
}