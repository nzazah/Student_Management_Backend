package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
)

type MongoReportMock struct {
	mock.Mock
}

func (m *MongoReportMock) SumPointsByIDs(ctx context.Context, ids []string) (int, error) {
	args := m.Called(ctx, ids)
	return args.Int(0), args.Error(1)
}

func (m *MongoReportMock) CountByType(ctx context.Context, ids []string) (map[string]int, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *MongoReportMock) FindByIDs(ctx context.Context, ids []string) ([]bson.M, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]bson.M), args.Error(1)
}