package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func NewFromEnv() (*sql.DB, error) {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

func EnsureSchema(conn *sql.DB) error {
	if conn == nil {
		return fmt.Errorf("database connection is nil")
	}
	_, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			article TEXT,
			supplier_article TEXT,
			sku TEXT,
			name TEXT,
			brand TEXT,
			category TEXT,
			price INTEGER,
			quantity INTEGER,
			stock INTEGER,
			data_issues TEXT[]
		);
	`)
	if err != nil {
		return err
	}
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS cart_items (
			session_id TEXT,
			sku TEXT,
			name TEXT,
			quantity INTEGER,
			price INTEGER
		);
	`)
	return err
}
