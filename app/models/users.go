package models

import "time"

type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	FullName     string
	RoleID       string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

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