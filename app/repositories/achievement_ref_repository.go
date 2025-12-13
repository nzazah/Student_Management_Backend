package repositories

import (
    "database/sql"
    "time"
    "uas/app/models"
    "github.com/lib/pq"
)

type IAchievementReferenceRepo interface {
    Create(ref *models.AchievementReference) (string, error)
    GetByID(id string) (*models.AchievementReference, error)
    UpdateStatusByMongoID(mongoID string, status string, submittedAt *time.Time) error
	SoftDeleteByMongoID(mongoID string) error
    GetByMongoID(mongoID string) (*models.AchievementReference, error)
    GetByStudentID(studentID string) ([]*models.AchievementReference, error)
    GetByStudentIDs(studentIDs []string) ([]*models.AchievementReference, error)
    GetAll() ([]*models.AchievementReference, error)
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

// Update status berdasarkan MongoAchievementID
func (r *AchievementReferenceRepo) UpdateStatusByMongoID(mongoID string, status string, submittedAt *time.Time) error {
    _, err := r.DB.Exec(`
        UPDATE achievement_references
        SET status=$2, submitted_at=$3, updated_at=NOW()
        WHERE mongo_achievement_id=$1
    `, mongoID, status, submittedAt)
    return err
}

// SoftDeleteByMongoID melakukan soft delete berdasarkan mongo_achievement_id
func (r *AchievementReferenceRepo) SoftDeleteByMongoID(mongoID string) error {
    now := time.Now() // gunakan untuk updated_at jika perlu
    _, err := r.DB.Exec(`
        UPDATE achievement_references
        SET status='deleted', updated_at=$2
        WHERE mongo_achievement_id=$1
    `, mongoID, now)
    return err
}




func (r *AchievementReferenceRepo) GetByStudentID(studentID string) ([]*models.AchievementReference, error) {
	rows, err := r.DB.Query(`
		SELECT id, student_id, mongo_achievement_id, status, created_at, updated_at
		FROM achievement_references
		WHERE student_id=$1 AND status<>'deleted'
	`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []*models.AchievementReference
	for rows.Next() {
		ref := &models.AchievementReference{}
		err := rows.Scan(&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status, &ref.CreatedAt, &ref.UpdatedAt)
		if err != nil {
			continue
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

func (r *AchievementReferenceRepo) GetByMongoID(mongoID string) (*models.AchievementReference, error) {
    ref := &models.AchievementReference{}
    err := r.DB.QueryRow(`
        SELECT id, student_id, mongo_achievement_id, status,
               submitted_at, verified_at, verified_by, rejection_note,
               created_at, updated_at
        FROM achievement_references
        WHERE mongo_achievement_id=$1
    `, mongoID).Scan(
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

func (r *AchievementReferenceRepo) GetByStudentIDs(studentIDs []string) ([]*models.AchievementReference, error) {
	query := `
		SELECT id, student_id, mongo_achievement_id, status, created_at, updated_at
		FROM achievement_references
		WHERE student_id = ANY($1) AND status <> 'deleted'
	`

	rows, err := r.DB.Query(query, pq.Array(studentIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []*models.AchievementReference
	for rows.Next() {
		ref := &models.AchievementReference{}
		if err := rows.Scan(
			&ref.ID,
			&ref.StudentID,
			&ref.MongoAchievementID,
			&ref.Status,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		); err != nil {
			continue
		}
		refs = append(refs, ref)
	}

	return refs, nil
}

func (r *AchievementReferenceRepo) GetAll() ([]*models.AchievementReference, error) {
	rows, err := r.DB.Query(`
		SELECT id, student_id, mongo_achievement_id, status,
		       submitted_at, created_at, updated_at
		FROM achievement_references
		WHERE status <> 'deleted'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []*models.AchievementReference
	for rows.Next() {
		ref := &models.AchievementReference{}
		err := rows.Scan(
			&ref.ID,
			&ref.StudentID,
			&ref.MongoAchievementID,
			&ref.Status,
			&ref.SubmittedAt,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		)
		if err != nil {
			continue
		}
		refs = append(refs, ref)
	}
	return refs, nil
}
