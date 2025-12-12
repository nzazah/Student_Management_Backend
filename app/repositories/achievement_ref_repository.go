package repositories

import (
    "database/sql"
    "time"
    "uas/app/models"
)

type IAchievementReferenceRepo interface {
    Create(ref *models.AchievementReference) (string, error)
    GetByID(id string) (*models.AchievementReference, error)
    UpdateStatus(id string, status string, submittedAt *time.Time) error
    SoftDelete(id string) error
}

type AchievementReferenceRepo struct {
    DB *sql.DB
}

func NewAchievementReferenceRepo(db *sql.DB) IAchievementReferenceRepo {
    return &AchievementReferenceRepo{DB: db}
}

func (r *AchievementReferenceRepo) Create(ref *models.AchievementReference) (string, error) {
    var id string
    err := r.DB.QueryRow(`
        INSERT INTO achievement_references 
        (student_id, mongo_achievement_id, status, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
        RETURNING id
    `, ref.StudentID, ref.MongoAchievementID, ref.Status).Scan(&id)

    return id, err
}

func (r *AchievementReferenceRepo) GetByID(id string) (*models.AchievementReference, error) {
    ref := &models.AchievementReference{}

    err := r.DB.QueryRow(`
        SELECT id, student_id, mongo_achievement_id, status,
               submitted_at, verified_at, verified_by, rejection_note,
               created_at, updated_at
        FROM achievement_references
        WHERE id=$1
    `, id).Scan(
        &ref.ID,
        &ref.StudentID,
        &ref.MongoAchievementID,
        &ref.Status,
        &ref.SubmittedAt,
        &ref.VerifiedAt,
        &ref.VerifiedBy,
        &ref.RejectionNote,
        &ref.CreatedAt,
        &ref.UpdatedAt,
    )

    return ref, err
}

func (r *AchievementReferenceRepo) UpdateStatus(id string, status string, submittedAt *time.Time) error {
    _, err := r.DB.Exec(`
        UPDATE achievement_references
        SET status=$2, submitted_at=$3, updated_at=NOW()
        WHERE id=$1
    `, id, status, submittedAt)
    return err
}

func (r *AchievementReferenceRepo) SoftDelete(id string) error {
    now := time.Now()
    _, err := r.DB.Exec(`
        UPDATE achievement_references
        SET status='deleted', verified_at=$2, updated_at=NOW()
        WHERE id=$1
    `, id, now)
    return err
}
