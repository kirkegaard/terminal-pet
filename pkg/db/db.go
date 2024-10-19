package db

import (
	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db     *sqlx.DB
	logger *log.Logger
}

func Open(driverName string, dsn string) (*DB, error) {
	db, err := sqlx.Connect(driverName, dsn)
	if err != nil {
		return nil, err
	}

	d := &DB{db: db}

	d.createUserTable()

	return d, nil
}

func (d *DB) Close() {
	d.db.Close()
}

func (d *DB) createUserTable() {
	schema := `
    CREATE TABLE IF NOT EXISTS users (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      public_key TEXT NOT NULL
    )
  `
	d.db.MustExec(schema)
}
