package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"pvz_service/internal/config"
)

type Storage struct {
	DB *sql.DB
}

func New(cfg *config.DatabaseConfig) (*Storage, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	return &Storage{DB: db}, nil
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.DB.PingContext(ctx)
}

func (s *Storage) Close() error {
	return s.DB.Close()
}
