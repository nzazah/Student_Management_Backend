package services

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"uas/app/models"
	"uas/app/repositories"
	"uas/databases"
	"uas/utils"
)

// Login godoc
// @Summary User login
// @Description Autentikasi user dan mengembalikan access token & refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body object true "Login credentials" example({"username":"john123","password":"password123"})
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
func Login(userRepo repositories.IUserRepository, refreshRepo repositories.IRefreshRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()

		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
		}

		user, err := userRepo.FindByUsername(ctx, req.Username)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "user not found"})
		}

		if !user.IsActive {
			return c.Status(403).JSON(fiber.Map{"error": "user inactive"})
		}

		if !utils.CheckPassword(req.Password, user.PasswordHash) {
			return c.Status(401).JSON(fiber.Map{"error": "invalid password"})
		}

		roleName, _ := userRepo.GetRoleName(ctx, user.RoleID)

		perms, err := userRepo.GetPermissionsByUserID(ctx, user.ID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to load permissions"})
		}

		accessToken, err := utils.GenerateAccessTokenWithPermissions(*user, roleName, perms)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to generate access token"})
		}

		refreshToken, err := utils.GenerateRefreshToken(user.ID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to generate refresh token"})
		}

		_ = refreshRepo.Delete(user.ID)
		err = refreshRepo.Save(user.ID, refreshToken)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to save refresh token"})
		}

		return c.JSON(fiber.Map{
			"token":        accessToken,
			"refreshToken": refreshToken,
			"user": fiber.Map{
				"id":          user.ID,
				"username":    user.Username,
				"fullName":    user.FullName,
				"role":        roleName,
				"permissions": perms,
			},
		})
	}
}

// Refresh godoc
// @Summary Refresh access token
// @Description Menghasilkan access token baru menggunakan refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refresh body object true "Refresh token" example({"refreshToken":"your_refresh_token"})
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /auth/refresh [post]
func Refresh(c *fiber.Ctx) error {
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

	userRepo := repositories.NewUserRepository(databases.PSQL)
	refreshRepo := repositories.NewRefreshRepository(databases.PSQL)

	if !refreshRepo.Exists(userID, req.RefreshToken) {
		return c.Status(401).JSON(fiber.Map{"error": "refresh token expired or revoked"})
	}

	user, err := userRepo.FindByID(ctx, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	roleName, _ := userRepo.GetRoleName(ctx, user.RoleID)
	perms, _ := userRepo.GetPermissionsByUserID(ctx, user.ID)

	newAccess, _ := utils.GenerateAccessTokenWithPermissions(*user, roleName, perms)
	newRefresh, _ := utils.GenerateRefreshToken(userID)

	_ = refreshRepo.Delete(userID)
	_ = refreshRepo.Save(userID, newRefresh)

	return c.JSON(fiber.Map{
		"token":        newAccess,
		"refreshToken": newRefresh,
	})
}

// Logout godoc
// @Summary Logout user
// @Description Menghapus refresh token dari database
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Security ApiKeyAuth
// @Router /auth/logout [post]
func Logout(c *fiber.Ctx) error {
	claims := c.Locals("user").(*models.JWTClaims)

	refreshRepo := repositories.NewRefreshRepository(databases.PSQL)
	_ = refreshRepo.Delete(claims.UserID)

	return c.JSON(fiber.Map{"message": "logged out"})
}

// Profile godoc
// @Summary Get user profile
// @Description Mendapatkan data profil user yang sedang login
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /auth/profile [get]
func Profile(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("user").(*models.JWTClaims)

	userRepo := repositories.NewUserRepository(databases.PSQL)
	studentRepo := repositories.NewStudentRepository(databases.PSQL)
	lecturerRepo := repositories.NewLecturerRepository(databases.PSQL)

	user, err := userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	roleName, _ := userRepo.GetRoleName(ctx, user.RoleID)

	result := fiber.Map{
		"id":          user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"fullName":    user.FullName,
		"role":        roleName,
		"permissions": claims.Permissions,
	}

	if roleName == "Mahasiswa" {
		if s, err := studentRepo.FindByUserID(user.ID); err == nil {
			result["mahasiswa"] = s
		}
	}

	if roleName == "Dosen Wali" {
		if d, err := lecturerRepo.FindByUserID(user.ID); err == nil {
			result["dosen_wali"] = d
		}
	}

	return c.JSON(fiber.Map{"user": result})
}
