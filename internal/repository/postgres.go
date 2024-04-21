package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	goodsTable    = "goods"
	projectsTable = "projects"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresDB(ctx context.Context, config Config) (*pgx.Conn, error) {
	const op = "storage.postgres.New"
	connString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", config.Host, config.Port, config.Username, config.DBName, config.Password, config.SSLMode)
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = conn.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return conn, nil
}
