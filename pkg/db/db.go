package db

import (
	"context"
	"sync"

	// "github.com/kirkegaard/terminal-pet/pkg/db/models"
	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sqlx.DB
	logger *log.Logger
}

var (
	instance *DB
	mu       sync.Mutex
)

func GetInstance() *DB {
	return instance
}

func SetInstance(db *DB) {
	mu.Lock()
	defer mu.Unlock()
	instance = db
}

func Open(ctx context.Context, driverName string, dsn string) (*DB, error) {
	db, err := sqlx.ConnectContext(ctx, driverName, dsn)
	if err != nil {
		return nil, err
	}

	d := &DB{DB: db}

	SetInstance(d)

	return d, nil
}

func (d *DB) Ping() error {
	return d.DB.Ping()
}

func (d *DB) Close() {
	d.DB.Close()
}

// Initializes the database schema
func (d *DB) CreateTables(ctx context.Context) error {
	_, err := d.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			public_key TEXT NOT NULL UNIQUE
		)
	`)
	if err != nil {
		return err
	}

	_, err = d.Exec(`
		CREATE TABLE IF NOT EXISTS pets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			birthday TIMESTAMP NOT NULL,
			parent_id INTEGER NOT NULL,
			hunger INTEGER NOT NULL DEFAULT 0,
			happiness INTEGER NOT NULL DEFAULT 0,
			discipline INTEGER NOT NULL DEFAULT 0,
			health INTEGER NOT NULL DEFAULT 100,
			weight INTEGER NOT NULL DEFAULT 0,
			is_sick BOOLEAN NOT NULL DEFAULT 0,
			has_pooped BOOLEAN NOT NULL DEFAULT 0,
			lights_on BOOLEAN NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (parent_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		return err
	}

	// // Check if the is_sick column exists
	// var count int
	// err = d.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('pets') WHERE name='is_sick'`).Scan(&count)
	// if err != nil {
	// 	return err
	// }

	// // Add is_sick column if it doesn't exist
	// if count == 0 {
	// 	_, err = d.Exec(`ALTER TABLE pets ADD COLUMN is_sick BOOLEAN NOT NULL DEFAULT 0`)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	log.Info("Added is_sick column to pets table")
	// }

	// // Check if the has_pooped column exists
	// err = d.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('pets') WHERE name='has_pooped'`).Scan(&count)
	// if err != nil {
	// 	return err
	// }

	// // Add has_pooped column if it doesn't exist
	// if count == 0 {
	// 	_, err = d.Exec(`ALTER TABLE pets ADD COLUMN has_pooped BOOLEAN NOT NULL DEFAULT 0`)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	log.Info("Added has_pooped column to pets table")
	// }

	log.Info("Database tables created successfully")
	return nil
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
