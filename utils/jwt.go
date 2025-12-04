package utils

import (
	"time"
	"uas/app/models"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("super-secret-key-should-be-long")
var refreshSecret = []byte("super-refresh-token-secret")

func GenerateToken(user models.User, role string, permissions []string) (string, error) {

	claims := models.JWTClaims{
		UserID:      user.ID,
		Username:    user.Username,
		Role:        role,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshSecret)
}
