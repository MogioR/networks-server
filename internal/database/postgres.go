package database

import (
	"database/sql"
	"fmt"
	"messager-server/internal/config"

	_ "github.com/lib/pq"
)

func ConnectToPostgres(cfg *config.Postgres) (*sql.DB, error) {
	DSN := fmt.Sprintf(
		"dbname=%s user=%s password=%s host=%s port=%d sslmode=%s",
		cfg.Db,
		cfg.User,
		cfg.Pass,
		cfg.Host,
		cfg.Port,
		cfg.Sslmode,
	)

	db, _ := sql.Open("postgres", DSN)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ConnectToDb: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxConns)

	return db, nil
}
