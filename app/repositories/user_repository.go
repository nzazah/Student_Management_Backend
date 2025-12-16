package repositories

import (
	"context"
	"database/sql"
	"uas/app/models"
)

type IUserRepository interface {
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	FindByID(ctx context.Context, userID string) (*models.User, error)

	GetRoleName(ctx context.Context, roleID string) (string, error)
	GetPermissionsByUserID(ctx context.Context, userID string) ([]string, error)
}

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}


func (r *UserRepository) GetPermissionsByUserID(ctx context.Context, userID string) ([]string, error) {
	const query = `
	SELECT p.name
	FROM permissions p
	JOIN role_permissions rp ON rp.permission_id = p.id
	JOIN roles r ON r.id = rp.role_id
	JOIN users u ON u.role_id = r.id
	WHERE u.id = $1;
	`
	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}
	return permissions, nil
}


func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
	SELECT id, username, email, password_hash, full_name, role_id, is_active
	FROM users
	WHERE username = $1 OR email = $1
	LIMIT 1
	`
	row := r.DB.QueryRowContext(ctx, query, username)

	var u models.User
	if err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.RoleID,
		&u.IsActive,
	); err != nil {
		return nil, err
	}
	return &u, nil
}


func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	query := `
	SELECT id, username, email, password_hash, full_name, role_id, is_active
	FROM users
	WHERE id = $1
	LIMIT 1
	`
	row := r.DB.QueryRowContext(ctx, query, id)

	var u models.User
	if err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.RoleID,
		&u.IsActive,
	); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetRoleName(ctx context.Context, roleID string) (string, error) {
	query := `SELECT name FROM roles WHERE id = $1`
	row := r.DB.QueryRowContext(ctx, query, roleID)

	var roleName string
	err := row.Scan(&roleName)
	return roleName, err
}
