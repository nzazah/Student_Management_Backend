package repositories

import (
	"database/sql"
	"uas/app/models"
)

type IUserRepository interface {
	FindByUsername(username string) (*models.User, error)
	FindByID(userID string) (*models.User, error)
	GetRoleName(roleID string) (string, error)
}

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) IUserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	query := `
        SELECT id, username, email, password_hash, full_name, role_id, is_active
        FROM users
        WHERE username = $1 OR email = $1
        LIMIT 1
    `
	row := r.DB.QueryRow(query, username)

	var u models.User
	err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.RoleID,
		&u.IsActive,
	)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) FindByID(id string) (*models.User, error) {
	query := `
        SELECT id, username, email, password_hash, full_name, role_id, is_active
        FROM users
        WHERE id = $1
        LIMIT 1
    `
	row := r.DB.QueryRow(query, id)

	var u models.User
	err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.RoleID,
		&u.IsActive,
	)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) GetRoleName(roleID string) (string, error) {
	query := `SELECT name FROM roles WHERE id = $1`
	row := r.DB.QueryRow(query, roleID)

	var roleName string
	err := row.Scan(&roleName)
	return roleName, err
}
