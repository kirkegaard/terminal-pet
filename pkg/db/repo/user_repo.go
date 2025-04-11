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

func (r *UserRepository) FindByPublicKey(ctx context.Context, publicKey string) (*models.User, error) {
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

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	_, err := r.db.Exec("INSERT INTO users (name, public_key) VALUES (?, ?)",
		user.Name,
		user.PublicKey,
	)

	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}
