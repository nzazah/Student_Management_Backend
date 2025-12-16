package repositories

import (
	"context"
	"uas/app/models"
)

type IUserManagementRepository interface {
	FindAll(ctx context.Context) ([]models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	UpdateRole(ctx context.Context, userID string, roleID string) error
}

func (r *UserRepository) FindAll(ctx context.Context) ([]models.User, error) {
	query := `
	SELECT id, username, email, full_name, role_id, is_active, created_at, updated_at
	FROM users
	ORDER BY created_at DESC
	`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.FullName,
			&u.RoleID,
			&u.IsActive,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
	INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.DB.ExecContext(
		ctx,
		query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.RoleID,
		user.IsActive,
	)
	return err
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
	UPDATE users
	SET username = $1, email = $2, full_name = $3, is_active = $4, updated_at = NOW()
	WHERE id = $5
	`
	_, err := r.DB.ExecContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.FullName,
		user.IsActive,
		user.ID,
	)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.DB.ExecContext(ctx, query, id)
	return err
}

func (r *UserRepository) UpdateRole(ctx context.Context, userID string, roleID string) error {
	query := `
	UPDATE users
	SET role_id = $1, updated_at = NOW()
	WHERE id = $2
	`
	_, err := r.DB.ExecContext(ctx, query, roleID, userID)
	return err
}
