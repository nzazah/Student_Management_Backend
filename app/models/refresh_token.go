package models

import "time"

type RefreshToken struct {
	ID        int       `json:"id"`
	UserID    string    `json:"userId"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"createdAt"`
}
