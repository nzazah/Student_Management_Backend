package repositories

import (
	"database/sql"
	"uas/app/models"
)

type IUserRepository interface {
	FindByUsername(username string) (*models.User, error)
	GetRoleName(roleID string) (string, error)
	GetPermissions(roleID string) ([]string, error)
}

type UserRepository struct {
	DB *sql.DB
}

// Factory untuk membuat repository
func NewUserRepository(db *sql.DB) IUserRepository {
	return &UserRepository{DB: db}
}

// -----------------------------------------------------------------------------
// FIND USER
// -----------------------------------------------------------------------------
func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User

	query := `
        SELECT id, username, email, password_hash, full_name, role_id, is_active
        FROM users
        WHERE username = $1 OR email = $1
        LIMIT 1
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

// -----------------------------------------------------------------------------
// GET ROLE NAME
// -----------------------------------------------------------------------------
func (r *UserRepository) GetRoleName(roleID string) (string, error) {
	var roleName string

	query := `
        SELECT name
        FROM roles
        WHERE id = $1
    `

	err := r.DB.QueryRow(query, roleID).Scan(&roleName)
	if err != nil {
		return "", err
	}

	return roleName, nil
}

// -----------------------------------------------------------------------------
// GET PERMISSIONS BY ROLE
// -----------------------------------------------------------------------------
func (r *UserRepository) GetPermissions(roleID string) ([]string, error) {
	query := `
        SELECT p.name
        FROM role_permissions rp
        JOIN permissions p ON p.id = rp.permission_id
        WHERE rp.role_id = $1
    `

	rows, err := r.DB.Query(query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err == nil {
			perms = append(perms, perm)
		}
	}

	return perms, nil
}
