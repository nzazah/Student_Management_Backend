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