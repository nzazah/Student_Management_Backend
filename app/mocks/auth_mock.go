package mocks

import (
	"context"
	"uas/app/models"
	"github.com/stretchr/testify/mock"
)

type UserRepoMock struct { mock.Mock }
func (m *UserRepoMock) FindByUsername(ctx context.Context, u string) (*models.User, error) {
	args := m.Called(ctx, u)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *UserRepoMock) GetRoleName(ctx context.Context, id string) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}
func (m *UserRepoMock) GetPermissionsByUserID(ctx context.Context, id string) ([]string, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]string), args.Error(1)
}


type RefreshRepoMock struct { mock.Mock }
func (m *RefreshRepoMock) Save(uid string, t string) error { return m.Called(uid, t).Error(0) }
func (m *RefreshRepoMock) Delete(uid string) error { return m.Called(uid).Error(0) }


func (m *UserRepoMock) Create(ctx context.Context, user *models.User) error {
    return m.Called(ctx, user).Error(0)
}

func (m *UserRepoMock) Update(ctx context.Context, user *models.User) error {
    return m.Called(ctx, user).Error(0)
}

func (m *UserRepoMock) Delete(ctx context.Context, id string) error {
    return m.Called(ctx, id).Error(0)
}

func (m *UserRepoMock) FindAll(ctx context.Context) ([]models.User, error) {
    args := m.Called(ctx)
    return args.Get(0).([]models.User), args.Error(1)
}

func (m *UserRepoMock) FindByID(ctx context.Context, id string) (*models.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil { return nil, args.Error(1) }
    return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserRepoMock) UpdateRole(ctx context.Context, userID string, roleID string) error {
    return m.Called(ctx, userID, roleID).Error(0)
}

func (m *RefreshRepoMock) Exists(userID string, token string) bool {
    args := m.Called(userID, token)
    return args.Bool(0)
}