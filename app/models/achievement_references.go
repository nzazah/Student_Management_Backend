package models

import "time"

type AchievementReference struct {
	ID                 string
	StudentID          string
	MongoAchievementID string
	Status             string      
	SubmittedAt        *time.Time  
	VerifiedAt         *time.Time 
	VerifiedBy         *string    
	RejectionNote      *string     
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
