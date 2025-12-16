package repositories

import (
	"database/sql"
	"uas/app/models"
)

type ILecturerRepository interface {
	FindByUserID(userID string) (*models.Lecturer, error)
	Create(lecturer *models.Lecturer) error
	Update(lecturer *models.Lecturer) error
}

type LecturerRepository struct {
	DB *sql.DB
}

func NewLecturerRepository(db *sql.DB) ILecturerRepository {
	return &LecturerRepository{DB: db}
}

func (r *LecturerRepository) FindByUserID(userID string) (*models.Lecturer, error) {
	query := `
        SELECT id, user_id, lecturer_id, department, created_at
        FROM lecturers
        WHERE user_id = $1
        LIMIT 1
    `
	row := r.DB.QueryRow(query, userID)

	var l models.Lecturer
	err := row.Scan(
		&l.ID,
		&l.UserID,
		&l.LecturerID,
		&l.Department,
		&l.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &l, nil
}

func (r *LecturerRepository) Create(l *models.Lecturer) error {
	query := `
	INSERT INTO lecturers (id, user_id, lecturer_id, department, created_at)
	VALUES ($1, $2, $3, $4, NOW())
	`
	_, err := r.DB.Exec(
		query,
		l.ID,
		l.UserID,
		l.LecturerID,
		l.Department,
	)
	return err
}

func (r *LecturerRepository) Update(l *models.Lecturer) error {
	query := `
	UPDATE lecturers
	SET
		lecturer_id = $1,
		department = $2
	WHERE user_id = $3
	`
	_, err := r.DB.Exec(
		query,
		l.LecturerID,
		l.Department,
		l.UserID,
	)
	return err
}
