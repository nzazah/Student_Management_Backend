package models

import "time"

type Student struct {
	ID           string
	UserID       string
	StudentID    string
	ProgramStudy string
	AcademicYear string
	AdvisorID    *string 
	CreatedAt    time.Time
}
