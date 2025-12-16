package services

import (
	"context"
	"time"
	"fmt"
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

	Student  *StudentPayload  `json:"student,omitempty"`
	Lecturer *LecturerPayload `json:"lecturer,omitempty"`
}

type StudentPayload struct {
	StudentID    string  `json:"studentId"`
	ProgramStudy string  `json:"programStudy"`
	AcademicYear string  `json:"academicYear"`
	AdvisorID    *string `json:"advisorId,omitempty"`
}

type LecturerPayload struct {
	LecturerID string `json:"lecturerId"`
	Department string `json:"department"`
}


type AssignRoleRequest struct {
	RoleID string `json:"role_id"`
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

	return c.JSON(fiber.Map{
		"user": user,
	})
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
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}

	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to hash password")
	}

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

	// create user
	if err := s.userRepo.Create(ctx, user); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// === create student profile jika RoleID mahasiswa ===
	if req.RoleID == "e8c7223b-2905-45c7-b7d0-8b06345dd1d7" && req.Student != nil {
		student := &models.Student{
			ID:           uuid.NewString(),
			UserID:       user.ID,
			StudentID:    req.Student.StudentID,
			ProgramStudy: req.Student.ProgramStudy,
			AcademicYear: req.Student.AcademicYear,
			AdvisorID:    req.Student.AdvisorID,
			CreatedAt:    time.Now(),
		}

		if err := s.studentRepo.Create(student); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to create student profile")
		}
	}

	// === create lecturer profile jika RoleID dosen ===
	if req.RoleID == "f09c60fc-abd0-45b2-930d-7be7e9f7599e" && req.Lecturer != nil {
		lecturer := &models.Lecturer{
			ID:         uuid.NewString(),
			UserID:     user.ID,
			LecturerID: req.Lecturer.LecturerID,
			Department: req.Lecturer.Department,
			CreatedAt:  time.Now(),
		}

		fmt.Printf("Creating lecturer: %+v\n", lecturer)

		if err := s.lecturerRepo.Create(lecturer); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to create lecturer profile")
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
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