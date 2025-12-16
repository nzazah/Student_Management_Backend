package repositories

import (
	"database/sql"
)

type ReportRepository interface {
	GetVerifiedAchievementMongoIDs() ([]string, error)
	GetVerifiedAchievementMongoIDsByStudent(studentID string) ([]string, error)
}

type reportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) GetVerifiedAchievementMongoIDs() ([]string, error) {
	rows, err := r.db.Query("SELECT mongo_achievement_id FROM achievement_references WHERE status = 'verified'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *reportRepository) GetVerifiedAchievementMongoIDsByStudent(studentID string) ([]string, error) {
	rows, err := r.db.Query("SELECT mongo_achievement_id FROM achievement_references WHERE student_id = $1 AND status = 'verified'", studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
