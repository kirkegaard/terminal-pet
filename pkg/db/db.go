package db

import (
	"context"

	// "github.com/kirkegaard/terminal-pet/pkg/db/models"

	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sqlx.DB
	logger *log.Logger
}

func Open(ctx context.Context, driverName string, dsn string) (*DB, error) {
	db, err := sqlx.ConnectContext(ctx, driverName, dsn)
	if err != nil {
		return nil, err
	}

	d := &DB{DB: db}

	d.createUserTable()

	return d, nil
}

func (d *DB) Close() {
	d.Close()
}

func (d *DB) createUserTable() {
	schema := `
    CREATE TABLE IF NOT EXISTS users (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      public_key TEXT NOT NULL
    )
  `
	d.MustExec(schema)
}

func (d *DB) FindUserByPublicKey(key string) (bool, error) {
	var count int
	err := d.Get(&count, "SELECT COUNT(*) FROM users WHERE public_key = ?", key)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (d *DB) CreateUser() error {
	log.Info("Creating user")
	return nil
}
