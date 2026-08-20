package config

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
)

func InitPostgres(ctx context.Context) (*pgx.Conn, error) {
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	return conn, nil
}
