package middleware

import (
	"context"
	"strings"

	"uas/app/models"
	"uas/app/repositories"

	"github.com/gofiber/fiber/v2"
)

func RequirePermission(required string, userRepo repositories.IUserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {

		userLocal := c.Locals("user")
		if userLocal == nil {
			return c.Status(401).JSON(fiber.Map{"error": "missing auth context"})
		}

		claims, ok := userLocal.(*models.JWTClaims)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "invalid auth context"})
		}

		ctx := context.Background()

		// Jika token tidak membawa permissions → load dari DB
		if len(claims.Permissions) == 0 {
			perms, err := userRepo.GetPermissionsByUserID(ctx, claims.UserID)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "failed to load permissions"})
			}
			claims.Permissions = perms
			c.Locals("user", claims)
		}

		// Cek permission
		for _, p := range claims.Permissions {
			if strings.EqualFold(p, required) {
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{
			"error": "forbidden: insufficient permissions",
		})
	}
}
