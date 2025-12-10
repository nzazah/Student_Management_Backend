package repositories

import (
	"database/sql"
	"log"
)

type IRefreshRepository interface {
	Save(userID string, token string) error
	Delete(userID string) error
	Exists(userID string, token string) bool
}

type RefreshRepository struct {
	DB *sql.DB
}

func NewRefreshRepository(db *sql.DB) IRefreshRepository {
	return &RefreshRepository{DB: db}
}

func (r *RefreshRepository) Save(userID string, token string) error {
    _, err := r.DB.Exec(`
        INSERT INTO refresh_tokens (user_id, token)
        VALUES ($1, $2)
    `, userID, token)

    if err != nil {
        log.Println("ERROR INSERT REFRESH TOKEN:", err)
    }

    return err
}


func (r *RefreshRepository) Delete(userID string) error {
	_, err := r.DB.Exec(`DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}

func (r *RefreshRepository) Exists(userID string, token string) bool {
	row := r.DB.QueryRow(`
		SELECT id FROM refresh_tokens
		WHERE user_id = $1 AND token = $2
		LIMIT 1
	`, userID, token)

	var id int
	err := row.Scan(&id)
	return err == nil
}
