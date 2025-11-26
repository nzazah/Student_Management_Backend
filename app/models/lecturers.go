package models

import "time"

type Lecturer struct {
	ID          string
	UserID      string
	LecturerID  string
	Department  string
	CreatedAt   time.Time
}
