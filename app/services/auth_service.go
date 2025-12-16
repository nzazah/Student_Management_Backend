package services

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"uas/app/models"
	"uas/app/repositories"
	"uas/utils"
)

type AuthService struct {
	RepoUser     repositories.IUserRepository
	RepoStudent  repositories.IStudentRepository
	RepoLecturer repositories.ILecturerRepository
	RepoRefresh  repositories.IRefreshRepository
}

func NewAuthService(
	userRepo repositories.IUserRepository,
	studentRepo repositories.IStudentRepository,
	lecturerRepo repositories.ILecturerRepository,
	refreshRepo repositories.IRefreshRepository,
) *AuthService {
	return &AuthService{
		RepoUser:     userRepo,
		RepoStudent:  studentRepo,
		RepoLecturer: lecturerRepo,
		RepoRefresh:  refreshRepo,
	}
}

/*
====================================================
LOGIN
====================================================
*/

func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx := context.Background()

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	user, err := s.RepoUser.FindByUsername(ctx, req.Username)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "user not found"})
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return c.Status(401).JSON(fiber.Map{"error": "invalid password"})
	}

	roleName, _ := s.RepoUser.GetRoleName(ctx, user.RoleID)

	permissions, err := s.RepoUser.GetPermissionsByUserID(ctx, user.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to load permissions"})
	}

	accessToken, err := utils.GenerateAccessTokenWithPermissions(*user, roleName, permissions)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate access token"})
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate refresh token"})
	}

	_ = s.RepoRefresh.Delete(user.ID)
	_ = s.RepoRefresh.Save(user.ID, refreshToken)

	return c.JSON(fiber.Map{
		"token":        accessToken,
		"refreshToken": refreshToken,
		"user": fiber.Map{
			"id":          user.ID,
			"username":    user.Username,
			"fullName":    user.FullName,
			"role":        roleName,
			"permissions": permissions,
		},
	})
}

/*
====================================================
REFRESH TOKEN
====================================================
*/

func (s *AuthService) Refresh(c *fiber.Ctx) error {
	ctx := context.Background()

	var req struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		return utils.RefreshSecret, nil
	})
	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "invalid refresh token"})
	}

	userID := claims.Subject

	if !s.RepoRefresh.Exists(userID, req.RefreshToken) {
		return c.Status(401).JSON(fiber.Map{"error": "refresh token expired or revoked"})
	}

	user, err := s.RepoUser.FindByID(ctx, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	roleName, _ := s.RepoUser.GetRoleName(ctx, user.RoleID)

	permissions, err := s.RepoUser.GetPermissionsByUserID(ctx, user.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to load permissions"})
	}

	newAccess, err := utils.GenerateAccessTokenWithPermissions(*user, roleName, permissions)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate access token"})
	}

	newRefresh, err := utils.GenerateRefreshToken(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate refresh token"})
	}

	_ = s.RepoRefresh.Delete(userID)
	_ = s.RepoRefresh.Save(userID, newRefresh)

	return c.JSON(fiber.Map{
		"token":        newAccess,
		"refreshToken": newRefresh,
	})
}

/*
====================================================
LOGOUT
====================================================
*/

func (s *AuthService) Logout(c *fiber.Ctx) error {
	claims := c.Locals("user").(*models.JWTClaims)
	_ = s.RepoRefresh.Delete(claims.UserID)

	return c.JSON(fiber.Map{"message": "logged out"})
}

/*
====================================================
PROFILE
====================================================
*/

func (s *AuthService) Profile(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("user").(*models.JWTClaims)

	user, err := s.RepoUser.FindByID(ctx, claims.UserID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	roleName, _ := s.RepoUser.GetRoleName(ctx, user.RoleID)

	result := fiber.Map{
		"id":          user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"fullName":    user.FullName,
		"role":        roleName,
		"createdAt":   user.CreatedAt,
		"permissions": claims.Permissions,
	}

	if roleName == "Mahasiswa" {
		if student, err := s.RepoStudent.FindByUserID(user.ID); err == nil {
			result["mahasiswa"] = student
		}
	}

	if roleName == "Dosen Wali" {
		if lecturer, err := s.RepoLecturer.FindByUserID(user.ID); err == nil {
			result["dosen_wali"] = lecturer
		}
	}

	return c.JSON(fiber.Map{"user": result})
}
