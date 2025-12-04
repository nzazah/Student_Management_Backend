package services

import (
	"github.com/gofiber/fiber/v2"
	"uas/app/repositories"
	"uas/utils"
)

type AuthService struct {
	Repo repositories.IUserRepository
}

func NewAuthService(repo repositories.IUserRepository) *AuthService {
	return &AuthService{Repo: repo}
}

func (s *AuthService) Login(c *fiber.Ctx) error {

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid JSON"})
	}

	user, err := s.Repo.FindByUsername(req.Username)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "user not found"})
	}

	if !user.IsActive {
		return c.Status(403).JSON(fiber.Map{"error": "user is inactive"})
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return c.Status(401).JSON(fiber.Map{"error": "wrong password"})
	}

	// Load Role Name
	roleName, _ := s.Repo.GetRoleName(user.RoleID)

	// Load Permissions
	permissions, _ := s.Repo.GetPermissions(user.RoleID)

	// Generate JWT
	token, err := utils.GenerateToken(*user, roleName, permissions)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "token generation failed"})
	}

	refresh, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "refresh token failed"})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"token":        token,
			"refreshToken": refresh,
			"user": fiber.Map{
				"id":          user.ID,
				"username":    user.Username,
				"fullName":    user.FullName,
				"role":        roleName,
				"permissions": permissions,
			},
		},
	})
}
