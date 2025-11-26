package services

import (
	"github.com/gofiber/fiber/v2"
	"uas/app/repositories"
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

	if req.Password != user.PasswordHash {
		return c.Status(401).JSON(fiber.Map{"error": "wrong password"})
	}

	return c.JSON(fiber.Map{
		"message": "login success",
		"user":    user,
	})
}
