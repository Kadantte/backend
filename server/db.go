package server

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB = pgxpool.Pool

func newDB() (*DB, error) {
	ctx := context.Background()

	dbHost := os.Getenv("MORBO_DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("MORBO_DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("MORBO_DB_USER")
	if dbUser == "" {
		dbUser = "morbo"
	}
	dbPassword := os.Getenv("MORBO_DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "morbo"
	}
	dbName := os.Getenv("MORBO_DB_NAME")
	if dbName == "" {
		dbName = "morbo"
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
