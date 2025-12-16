package repositories

import (
	"database/sql"
	"uas/app/models"
)

type IStudentRepository interface {
	FindByUserID(userID string) (*models.Student, error)
	FindByAdvisorID(advisorID string) ([]*models.Student, error)
	Create(student *models.Student) error
	UpdateAdvisor(studentID string, advisorID string) error
	FindAll() ([]*models.Student, error)
	FindByID(studentID string) (*models.Student, error)
	FindAchievementsByStudentID(studentID string) ([]map[string]any, error)
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

func (r *StudentRepository) Create(s *models.Student) error {
	query := `
		INSERT INTO students (
			id, user_id, student_id,
			program_study, academic_year,
			advisor_id, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,NOW())
	`
	_, err := r.DB.Exec(
		query,
		s.ID,
		s.UserID,
		s.StudentID,
		s.ProgramStudy,
		s.AcademicYear,
		s.AdvisorID,
	)
	return err
}

func (r *StudentRepository) UpdateAdvisor(studentID string, advisorID string) error {
	query := `
	UPDATE students
	SET advisor_id = $1
	WHERE id = $2
	`
	_, err := r.DB.Exec(query, advisorID, studentID)
	return err
}

func (r *StudentRepository) FindAll() ([]*models.Student, error) {
	rows, err := r.DB.Query(`
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []*models.Student
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.StudentID,
			&s.ProgramStudy,
			&s.AcademicYear,
			&s.AdvisorID,
			&s.CreatedAt,
		); err != nil {
			return nil, err
		}
		students = append(students, &s)
	}
	return students, nil
}

func (r *StudentRepository) FindByID(studentID string) (*models.Student, error) {
	var s models.Student
	err := r.DB.QueryRow(`
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		WHERE id = $1
	`, studentID).Scan(
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

func (r *StudentRepository) FindAchievementsByStudentID(studentID string) ([]map[string]any, error) {
	rows, err := r.DB.Query(`
		SELECT id, title, description, status, created_at
		FROM achievements
		WHERE student_id = $1
	`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []map[string]any
	for rows.Next() {
		var (
			id, title, description, status string
			createdAt                     any
		)

		if err := rows.Scan(&id, &title, &description, &status, &createdAt); err != nil {
			return nil, err
		}

		achievements = append(achievements, map[string]any{
			"id":          id,
			"title":       title,
			"description": description,
			"status":      status,
			"created_at":  createdAt,
		})
	}
	return achievements, nil
}
