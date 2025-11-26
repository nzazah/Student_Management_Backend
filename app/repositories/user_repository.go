package repositories

import (
	"database/sql"
	"uas/app/models"
)


type IUserRepository interface {
	FindByUsername(username string) (*models.User, error)
}


type UserRepository struct {
	DB *sql.DB
}

// Factory untuk membuat repository
func NewUserRepository(db *sql.DB) IUserRepository {
	return &UserRepository{DB: db}
}

// Implementasi function dari interface
func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User

	query := `
        SELECT id, username, email, password_hash, full_name, role_id, is_active
        FROM users
        WHERE username = $1 LIMIT 1
    `

	err := r.DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.RoleID,
		&user.IsActive,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
