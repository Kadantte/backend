package db

import "github.com/jackc/pgx/v5/pgxpool"

type DB struct {
	pool *pgxpool.Pool
}
