package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func NewSQLite(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func Save(db *sql.DB, url string) error {
	_, err := db.Exec(`
		INSERT OR IGNORE INTO profiles (url)
		VALUES (?)
	`, url)

	return err
}
