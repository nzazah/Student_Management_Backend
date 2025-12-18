package services

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"uas/app/models"
	"uas/app/repositories"
	"uas/databases"
)

// GetAllUsers godoc
// @Summary Get all users
// @Description Mengambil semua data user
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /users [get]
func GetAllUsers(c *fiber.Ctx) error {
	ctx := context.Background()

	userRepo := repositories.NewUserRepository(databases.PSQL)

	users, err := userRepo.FindAll(ctx)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(fiber.Map{"data": users})
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Mengambil data user berdasarkan ID
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /users/{id} [get]
func GetUserByID(c *fiber.Ctx) error {
	ctx := context.Background()
	id := c.Params("id")

	userRepo := repositories.NewUserRepository(databases.PSQL)

	user, err := userRepo.FindByID(ctx, id)
	if err != nil {
		return fiber.NewError(404, "user not found")
	}

	return c.JSON(fiber.Map{"user": user})
}

// CreateUser godoc
// @Summary Create new user
// @Description Membuat user baru, bisa sekaligus mahasiswa atau dosen
// @Tags User
// @Accept json
// @Produce json
// @Param body body models.CreateUserRequest true "User payload"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /users [post]
func CreateUser(c *fiber.Ctx) error {
	ctx := context.Background()

	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "invalid request")
	}

	userRepo := repositories.NewUserRepository(databases.PSQL)
	studentRepo := repositories.NewStudentRepository(databases.PSQL)
	lecturerRepo := repositories.NewLecturerRepository(databases.PSQL)

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	user := &models.User{
		ID:           uuid.NewString(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		FullName:     req.FullName,
		RoleID:       req.RoleID,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := userRepo.Create(ctx, user); err != nil {
		return fiber.NewError(500, err.Error())
	}

	if req.Student != nil {
		studentRepo.Create(&models.Student{
			ID:           uuid.NewString(),
			UserID:       user.ID,
			StudentID:    req.Student.StudentID,
			ProgramStudy: req.Student.ProgramStudy,
			AcademicYear: req.Student.AcademicYear,
			AdvisorID:    req.Student.AdvisorID,
			CreatedAt:    time.Now(),
		})
	}

	if req.Lecturer != nil {
		lecturerRepo.Create(&models.Lecturer{
			ID:         uuid.NewString(),
			UserID:     user.ID,
			LecturerID: req.Lecturer.LecturerID,
			Department: req.Lecturer.Department,
			CreatedAt:  time.Now(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "user created",
		"user_id": user.ID,
	})
}


// UpdateUser godoc
// @Summary Update user
// @Description Memperbarui data user
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body map[string]interface{} true "User payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /users/{id} [put]
func UpdateUser(c *fiber.Ctx) error {
	ctx := context.Background()
	id := c.Params("id")

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "invalid request")
	}

	userRepo := repositories.NewUserRepository(databases.PSQL)

	if err := userRepo.Update(ctx, &models.User{
		ID:       id,
		Username: req.Username,
		Email:    req.Email,
		FullName: req.FullName,
		IsActive: req.IsActive,
	}); err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(fiber.Map{"message": "user updated"})
}

// DeleteUser godoc
// @Summary Delete user
// @Description Menghapus user berdasarkan ID
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /users/{id} [delete]
func DeleteUser(c *fiber.Ctx) error {
	ctx := context.Background()
	id := c.Params("id")

	userRepo := repositories.NewUserRepository(databases.PSQL)

	if err := userRepo.Delete(ctx, id); err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(fiber.Map{"message": "user deleted"})
}

// AssignRole godoc
// @Summary Assign role to user
// @Description Mengubah role user
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body models.AssignRoleRequest true "Role payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /users/{id}/role [put]
func AssignRole(c *fiber.Ctx) error {
	ctx := context.Background()
	id := c.Params("id")

	var req models.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "invalid request")
	}

	userRepo := repositories.NewUserRepository(databases.PSQL)

	if err := userRepo.UpdateRole(ctx, id, req.RoleID); err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(fiber.Map{"message": "role updated"})
}

