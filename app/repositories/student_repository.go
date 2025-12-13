package repositories

import (
	"database/sql"
	"uas/app/models"
)

type IStudentRepository interface {
	FindByUserID(userID string) (*models.Student, error)
	FindByAdvisorID(advisorID string) ([]*models.Student, error)

}

type StudentRepository struct {
	DB *sql.DB
}

func NewStudentRepository(db *sql.DB) IStudentRepository {
	return &StudentRepository{DB: db}
}

func (r *StudentRepository) FindByUserID(userID string) (*models.Student, error) {
	query := `
        SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
        FROM students
        WHERE user_id = $1
        LIMIT 1
    `
	row := r.DB.QueryRow(query, userID)

	var s models.Student
	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.StudentID,
		&s.ProgramStudy,
		&s.AcademicYear,
		&s.AdvisorID,
		&s.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *StudentRepository) FindByAdvisorID(advisorID string) ([]*models.Student, error) {
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		WHERE advisor_id = $1
	`

	rows, err := r.DB.Query(query, advisorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []*models.Student

	for rows.Next() {
		var s models.Student
		err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.StudentID,
			&s.ProgramStudy,
			&s.AcademicYear,
			&s.AdvisorID,
			&s.CreatedAt,
		)
		if err != nil {
			continue
		}

		students = append(students, &s)
	}

	return students, nil
}
