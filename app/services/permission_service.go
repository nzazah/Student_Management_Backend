package services

import (
	"context"
	"uas/app/repositories"
)

type PermissionService interface {
	GetPermissions(ctx context.Context, userID string) ([]string, error)
}

type permissionService struct {
	userRepo repositories.IUserRepository
}

func NewPermissionService(userRepo repositories.IUserRepository) PermissionService {
	return &permissionService{
		userRepo: userRepo,
	}
}

func (s *permissionService) GetPermissions(
	ctx context.Context,
	userID string,
) ([]string, error) {
	return s.userRepo.GetPermissionsByUserID(ctx, userID)
}
