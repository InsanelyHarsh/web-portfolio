package config

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/jackc/pgx/v5"
)

func InitPostgres(ctx context.Context) (*pgx.Conn, error) {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	if host == "" || user == "" || password == "" || dbname == "" {
		return nil, errors.New("DB_HOST, DB_USER, DB_PASSWORD, and DB_NAME must be set")
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	sslmode := os.Getenv("DB_SSLMODE")
	// if sslmode == "" {
	// 	sslmode = "require"
	// }

	connURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   "/" + dbname,
	}
	q := connURL.Query()
	q.Set("sslmode", sslmode)
	connURL.RawQuery = q.Encode()

	conn, err := pgx.Connect(ctx, connURL.String())
	if err != nil {
		return nil, err
	}

	return conn, nil
}
