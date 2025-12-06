package services

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"uas/app/repositories"
	"uas/utils"
	"uas/app/models"
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
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	user, err := s.Repo.FindByUsername(req.Username)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "user not found"})
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return c.Status(401).JSON(fiber.Map{"error": "invalid password"})
	}

	roleName, _ := s.Repo.GetRoleName(user.RoleID)

	access, _ := utils.GenerateAccessToken(*user, roleName)
	refresh, _ := utils.GenerateRefreshToken(user.ID)

	return c.JSON(fiber.Map{
		"token":        access,
		"refreshToken": refresh,
	})
}

func (s *AuthService) Refresh(c *fiber.Ctx) error {

	var req struct {
		Refresh string `json:"refreshToken"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(req.Refresh, claims, func(t *jwt.Token) (interface{}, error) {
		return utils.RefreshSecret, nil
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "invalid refresh token"})
	}

	user, _ := s.Repo.FindByID(claims.Subject)
	roleName, _ := s.Repo.GetRoleName(user.RoleID)

	access, _ := utils.GenerateAccessToken(*user, roleName)
	refresh, _ := utils.GenerateRefreshToken(user.ID)

	return c.JSON(fiber.Map{
		"token":        access,
		"refreshToken": refresh,
	})
}

func (s *AuthService) Logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "logged out"})
}

func (s *AuthService) Profile(c *fiber.Ctx) error {

	claims := c.Locals("user").(*models.JWTClaims)

	user, err := s.Repo.FindByID(claims.UserID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	roleName, _ := s.Repo.GetRoleName(user.RoleID)

	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"fullName":  user.FullName,
			"role":      roleName,
			"isActive":  user.IsActive,
			"createdAt": user.CreatedAt,
		},
	})
}

