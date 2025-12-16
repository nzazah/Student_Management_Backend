package services

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"uas/app/models"
	"uas/app/repositories"
)

/*
====================================================
INTERFACE
====================================================
*/

type IUserService interface {
	GetAllUsers(c *fiber.Ctx) error
	GetUserByID(c *fiber.Ctx) error
	CreateUser(c *fiber.Ctx) error
	UpdateUser(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx) error
	AssignRole(c *fiber.Ctx) error
	SetAdvisor(c *fiber.Ctx) error
}

/*
====================================================
SERVICE STRUCT
====================================================
*/

type UserService struct {
	userRepo     repositories.IUserManagementRepository
	studentRepo  repositories.IStudentRepository
	lecturerRepo repositories.ILecturerRepository
}

/*
====================================================
CONSTRUCTOR
====================================================
*/

func NewUserService(
	userRepo repositories.IUserManagementRepository,
	studentRepo repositories.IStudentRepository,
	lecturerRepo repositories.ILecturerRepository,
) IUserService {
	return &UserService{
		userRepo:     userRepo,
		studentRepo: studentRepo,
		lecturerRepo: lecturerRepo,
	}
}

/*
====================================================
DTO (REQUEST PAYLOAD)
====================================================
*/

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"fullName"`
	RoleID   string `json:"roleId"`

	Student  *models.Student  `json:"student,omitempty"`
	Lecturer *models.Lecturer `json:"lecturer,omitempty"`
}


type AssignRoleRequest struct {
	RoleID string `json:"role_id"`
}

type SetAdvisorRequest struct {
	AdvisorID string `json:"advisor_id"`
}

/*
====================================================
SERVICES
====================================================
*/

func (s *UserService) GetAllUsers(c *fiber.Ctx) error {
	ctx := context.Background()

	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(fiber.Map{"data": users})
}

func (s *UserService) GetUserByID(c *fiber.Ctx) error {
	ctx := context.Background()
	id := c.Params("id")

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return fiber.NewError(404, "user not found")
	}

	response := fiber.Map{"user": user}

	if student, err := s.studentRepo.FindByUserID(id); err == nil {
		response["student"] = student
	}

	if lecturer, err := s.lecturerRepo.FindByUserID(id); err == nil {
		response["lecturer"] = lecturer
	}

	return c.JSON(response)
}

/*
====================================================
CREATE USER (ADMIN)
====================================================
*/

func (s *UserService) CreateUser(c *fiber.Ctx) error {
	ctx := context.Background()

	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "invalid request")
	}

	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fiber.NewError(500, "failed to hash password")
	}

	user := &models.User{
		ID:           uuid.NewString(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		FullName:     req.FullName,
		RoleID:       req.RoleID,
		IsActive:     true,
	}

	// create user
	if err := s.userRepo.Create(ctx, user); err != nil {
		return fiber.NewError(500, err.Error())
	}

	// create student profile (optional)
	if req.Student != nil {
		req.Student.ID = uuid.NewString()
		req.Student.UserID = user.ID

		if err := s.studentRepo.Create(req.Student); err != nil {
			return fiber.NewError(500, "failed to create student profile")
		}
	}

	// create lecturer profile (optional)
	if req.Lecturer != nil {
		req.Lecturer.ID = uuid.NewString()
		req.Lecturer.UserID = user.ID

		if err := s.lecturerRepo.Create(req.Lecturer); err != nil {
			return fiber.NewError(500, "failed to create lecturer profile")
		}
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "user created",
		"user_id": user.ID,
	})
}

/*
====================================================
UPDATE USER (ADMIN)
====================================================
*/

func (s *UserService) UpdateUser(c *fiber.Ctx) error {
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

	user := &models.User{
		ID:       id,
		Username: req.Username,
		Email:    req.Email,
		FullName: req.FullName,
		IsActive: req.IsActive,
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(fiber.Map{"message": "user updated"})
}

/*
====================================================
DELETE USER (ADMIN)
====================================================
*/

func (s *UserService) DeleteUser(c *fiber.Ctx) error {
	ctx := context.Background()
	id := c.Params("id")

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(fiber.Map{"message": "user deleted"})
}

/*
====================================================
ASSIGN ROLE (ADMIN)
====================================================
*/

func (s *UserService) AssignRole(c *fiber.Ctx) error {
	ctx := context.Background()
	id := c.Params("id")

	var req AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "invalid request")
	}

	if err := s.userRepo.UpdateRole(ctx, id, req.RoleID); err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(fiber.Map{"message": "role updated"})
}

/*
====================================================
SET ADVISOR (ADMIN)
====================================================
*/

func (s *UserService) SetAdvisor(c *fiber.Ctx) error {
	studentID := c.Params("id")

	var req SetAdvisorRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "invalid request")
	}

	if err := s.studentRepo.UpdateAdvisor(studentID, req.AdvisorID); err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(fiber.Map{"message": "advisor assigned"})
}
