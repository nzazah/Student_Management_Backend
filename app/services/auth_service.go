package services

import (
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


func (s *AuthService) Login(c *fiber.Ctx) error {

    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }

    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
    }

    user, err := s.RepoUser.FindByUsername(req.Username)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": "user not found"})
    }

    if !utils.CheckPassword(req.Password, user.PasswordHash) {
        return c.Status(401).JSON(fiber.Map{"error": "invalid password"})
    }

    roleName, _ := s.RepoUser.GetRoleName(user.RoleID)

    access, _ := utils.GenerateAccessToken(*user, roleName)
    refresh, _ := utils.GenerateRefreshToken(user.ID)

    // Hapus refresh token lama
    _ = s.RepoRefresh.Delete(user.ID)

    // Simpan refresh token baru
    _ = s.RepoRefresh.Save(user.ID, refresh)

    return c.JSON(fiber.Map{
        "token":        access,
        "refreshToken": refresh,
    })
}



func (s *AuthService) Refresh(c *fiber.Ctx) error {
	var req struct {
		Refresh string `json:"refreshToken"`
	}
	c.BodyParser(&req)

	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(req.Refresh, claims, func(t *jwt.Token) (interface{}, error) {
		return utils.RefreshSecret, nil
	})
	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "invalid refresh token"})
	}

	userID := claims.Subject

	if !s.RepoRefresh.Exists(userID, req.Refresh) {
		return c.Status(401).JSON(fiber.Map{"error": "refresh token expired or revoked"})
	}

	user, _ := s.RepoUser.FindByID(userID)
	roleName, _ := s.RepoUser.GetRoleName(user.RoleID)

	newAccess, _ := utils.GenerateAccessToken(*user, roleName)
	newRefresh, _ := utils.GenerateRefreshToken(userID)

	_ = s.RepoRefresh.Delete(userID)
	_ = s.RepoRefresh.Save(userID, newRefresh)

	return c.JSON(fiber.Map{
		"token":        newAccess,
		"refreshToken": newRefresh,
	})
}



func (s *AuthService) Logout(c *fiber.Ctx) error {

	claims := c.Locals("user").(*models.JWTClaims)

	_ = s.RepoRefresh.Delete(claims.UserID)

	return c.JSON(fiber.Map{"message": "logged out"})
}



func (s *AuthService) Profile(c *fiber.Ctx) error {

	claims := c.Locals("user").(*models.JWTClaims)

	user, err := s.RepoUser.FindByID(claims.UserID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	roleName, _ := s.RepoUser.GetRoleName(user.RoleID)


	result := fiber.Map{
		"username":  user.Username,
		"email":     user.Email,
		"fullName":  user.FullName,
		"role":      roleName,
		"createdAt": user.CreatedAt,
	}


	if roleName == "Mahasiswa" {
		student, err := s.RepoStudent.FindByUserID(user.ID)
		if err == nil {
			result["Mahasiswa"] = fiber.Map{
				"studentId":    student.StudentID,
				"programStudy": student.ProgramStudy,
				"academicYear": student.AcademicYear,
			}
		}
	}


	if roleName == "Dosen Wali" {
		lecturer, err := s.RepoLecturer.FindByUserID(user.ID)
		if err == nil {
			result["Dosen Wali"] = fiber.Map{
				"lecturerId": lecturer.LecturerID,
				"department": lecturer.Department,
			}
		}
	}


	return c.JSON(fiber.Map{"user": result})
}
