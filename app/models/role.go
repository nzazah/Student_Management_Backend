package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Role struct {
	ID        string
	Name      string
	Description string
	CreatedAt time.Time
}

type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}