package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementAttachment struct {
	FileName   string    `bson:"fileName"`
	FileUrl    string    `bson:"fileUrl"`
	FileType   string    `bson:"fileType"`
	UploadedAt time.Time `bson:"uploadedAt"`
}

type AchievementDetails struct {
	CompetitionName  *string  `bson:"competitionName,omitempty"`
	CompetitionLevel *string  `bson:"competitionLevel,omitempty"`
	Rank             *int     `bson:"rank,omitempty"`
	MedalType        *string  `bson:"medalType,omitempty"`

	PublicationType  *string  `bson:"publicationType,omitempty"`
	PublicationTitle *string  `bson:"publicationTitle,omitempty"`
	Authors          []string `bson:"authors,omitempty"`
	Publisher        *string  `bson:"publisher,omitempty"`
	ISSN             *string  `bson:"issn,omitempty"`

	OrganizationName *string `bson:"organizationName,omitempty"`
	Position         *string `bson:"position,omitempty"`
	Period           *struct {
		Start time.Time `bson:"start"`
		End   time.Time `bson:"end"`
	} `bson:"period,omitempty"`

	CertificationName   *string    `bson:"certificationName,omitempty"`
	IssuedBy            *string    `bson:"issuedBy,omitempty"`
	CertificationNumber *string    `bson:"certificationNumber,omitempty"`
	ValidUntil          *time.Time `bson:"validUntil,omitempty"`

	EventDate *time.Time `bson:"eventDate,omitempty"`
	Location  *string    `bson:"location,omitempty"`
	Organizer *string    `bson:"organizer,omitempty"`
	Score     *float64   `bson:"score,omitempty"`

	CustomFields map[string]interface{} `bson:"customFields,omitempty"`
}

type MongoAchievement struct {
    ID              primitive.ObjectID      `bson:"_id,omitempty" json:"id"`
    StudentID       string                  `bson:"studentId" json:"studentId"`
    AchievementType string                  `bson:"achievementType" json:"achievementType"`
    Title           string                  `bson:"title" json:"title"`
    Description     string                  `bson:"description" json:"description"`
    Details         AchievementDetails      `bson:"details" json:"details"`
    Attachments     []AchievementAttachment `bson:"attachments" json:"attachments"`
    Tags            []string                `bson:"tags" json:"tags"`
    Points          int                     `bson:"points" json:"points"`
    CreatedAt       time.Time               `bson:"createdAt" json:"createdAt"`
    UpdatedAt       time.Time               `bson:"updatedAt" json:"updatedAt"`
    DeletedAt       *time.Time              `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`
}
