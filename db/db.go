package db

import (
	"morbo/log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
	log  log.Log
}

func NewDB() *DB {
	return &DB{nil, log.NewLog("db")}
}
