package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kirkegaard/terminal-pet/pkg/db"
	"github.com/kirkegaard/terminal-pet/pkg/db/models"
)

type UserRepository struct {
	db *db.DB
}

func NewUserRepository(database *db.DB) *UserRepository {
	if database == nil {
		database = db.GetInstance()
	}

	return &UserRepository{
		db: database,
	}
}

// GetByPublicKey retrieves a user ID by public key
func (r *UserRepository) GetByPublicKey(ctx context.Context, publicKey string) (int, error) {
	if r.db == nil {
		return 0, fmt.Errorf("no database connection available")
	}

	var userID int
	err := r.db.QueryRow("SELECT id FROM users WHERE public_key = ?", publicKey).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("find user by public key: %w", err)
	}

	return userID, nil
}

// Create creates a new user in the database
func (r *UserRepository) Create(ctx context.Context, name, publicKey string) (int, error) {
	if r.db == nil {
		return 0, fmt.Errorf("no database connection available")
	}

	res, err := r.db.Exec("INSERT INTO users (name, public_key) VALUES (?, ?)", name, publicKey)
	if err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}

	return int(id), nil
}

func (r *UserRepository) FindByPublicKey(ctx context.Context, publicKey string) (*models.User, error) {
	if r.db == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	var user models.User

	err := r.db.QueryRow("SELECT id, name, public_key FROM users WHERE public_key = ?", publicKey).Scan(
		&user.ID,
		&user.Name,
		&user.PublicKey,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by public key: %w", err)
	}

	return &user, nil
}
